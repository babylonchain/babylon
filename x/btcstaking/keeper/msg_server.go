package keeper

import (
	"context"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}
	if err := req.Params.Validate(); err != nil {
		return nil, govtypes.ErrInvalidProposalMsg.Wrapf("invalid parameter: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// CreateFinalityProvider creates a finality provider
func (ms msgServer) CreateFinalityProvider(goCtx context.Context, req *types.MsgCreateFinalityProvider) (*types.MsgCreateFinalityProviderResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricsKeyCreateFinalityProvider)

	// ensure the finality provider address does not already exist
	ctx := sdk.UnwrapSDKContext(goCtx)
	// basic stateless checks
	if err := req.ValidateBasic(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// verify proof of possession
	if err := req.Pop.Verify(req.BabylonPk, req.BtcPk, ms.btcNet); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid proof of possession: %v", err)
	}

	// ensure commission rate is
	// - at least the minimum commission rate in parameters, and
	// - at most 1
	if req.Commission.LT(ms.MinCommissionRate(ctx)) {
		return nil, types.ErrCommissionLTMinRate.Wrapf("cannot set finality provider commission to less than minimum rate of %s", ms.MinCommissionRate(ctx))
	}
	if req.Commission.GT(sdkmath.LegacyOneDec()) {
		return nil, types.ErrCommissionGTMaxRate
	}

	// ensure finality provider does not already exist
	if ms.HasFinalityProvider(ctx, *req.BtcPk) {
		return nil, types.ErrFpRegistered
	}

	// all good, add this finality provider
	fp := types.FinalityProvider{
		Description: req.Description,
		Commission:  req.Commission,
		BabylonPk:   req.BabylonPk,
		BtcPk:       req.BtcPk,
		Pop:         req.Pop,
	}
	ms.SetFinalityProvider(ctx, &fp)

	// notify subscriber
	if err := ctx.EventManager().EmitTypedEvent(&types.EventNewFinalityProvider{Fp: &fp}); err != nil {
		return nil, err
	}

	return &types.MsgCreateFinalityProviderResponse{}, nil
}

// EditFinalityProvider edits an existing finality provider
func (ms msgServer) EditFinalityProvider(ctx context.Context, req *types.MsgEditFinalityProvider) (*types.MsgEditFinalityProviderResponse, error) {
	// basic stateless checks
	// NOTE: after this, description is guaranteed to be valid
	if err := req.ValidateBasic(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// ensure commission rate is
	// - at least the minimum commission rate in parameters, and
	// - at most 1
	if req.Commission.LT(ms.MinCommissionRate(ctx)) {
		return nil, types.ErrCommissionLTMinRate.Wrapf("cannot set finality provider commission to less than minimum rate of %s", ms.MinCommissionRate(ctx))
	}
	if req.Commission.GT(sdkmath.LegacyOneDec()) {
		return nil, types.ErrCommissionGTMaxRate
	}

	// find the finality provider with the given BTC PK
	fp, err := ms.GetFinalityProvider(ctx, req.BtcPk)
	if err != nil {
		return nil, err
	}

	// ensure the signer corresponds to the finality provider's Babylon address
	fpBabylonAddr := sdk.AccAddress(fp.BabylonPk.Address())
	if req.Signer != fpBabylonAddr.String() {
		return nil, status.Errorf(codes.PermissionDenied, "the signer does not correspond to the finality provider's Babylon address")
	}

	// all good, update the finality provider and set back
	fp.Description = req.Description
	fp.Commission = req.Commission
	ms.SetFinalityProvider(ctx, fp)

	return &types.MsgEditFinalityProviderResponse{}, nil
}

// caluculateMinimumUnbondingValue calculates minimum unbonding value basend on current staking output value
// and params.MinUnbondingRate
func caluculateMinimumUnbondingValue(
	stakingOutput *wire.TxOut,
	params *types.Params,
) btcutil.Amount {
	// this conversions must always succeed, as it is part of our params
	minUnbondingRate := params.MinUnbondingRate.MustFloat64()
	// Caluclate min unbonding output value based on staking output, use btc native multiplication
	minUnbondingOutputValue := btcutil.Amount(stakingOutput.Value).MulF64(minUnbondingRate)
	return minUnbondingOutputValue
}

// CreateBTCDelegation creates a BTC delegation
// TODO: refactor this handler. It's now too convoluted
func (ms msgServer) CreateBTCDelegation(goCtx context.Context, req *types.MsgCreateBTCDelegation) (*types.MsgCreateBTCDelegationResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricsKeyCreateBTCDelegation)

	ctx := sdk.UnwrapSDKContext(goCtx)
	// basic stateless checks
	if err := req.ValidateBasic(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	vp := ms.GetParamsWithVersion(ctx)
	btccParams := ms.btccKeeper.GetParams(ctx)
	kValue, wValue := btccParams.BtcConfirmationDepth, btccParams.CheckpointFinalizationTimeout

	minUnbondingTime := types.MinimumUnbondingTime(vp.Params, btccParams)

	// Check unbonding time (staking time from unbonding tx) is larger than min unbonding time
	// which is larger value from:
	// - MinUnbondingTime
	// - CheckpointFinalizationTimeout
	if uint64(req.UnbondingTime) <= minUnbondingTime {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding time %d must be larger than %d", req.UnbondingTime, minUnbondingTime)
	}

	// At this point we know that unbonding time in request:
	// - is larger than min unbonding time
	// - is smaller than math.MaxUint16 (due to check in req.ValidateBasic())
	validatedUnbondingTime := uint16(req.UnbondingTime)

	// verify proof of possession
	if err := req.Pop.Verify(req.BabylonPk, req.BtcPk, ms.btcNet); err != nil {
		return nil, types.ErrInvalidProofOfPossession.Wrapf("error while validating proof of posession: %v", err)
	}

	// Ensure all finality providers are known to Babylon
	for _, fpBTCPK := range req.FpBtcPkList {
		if !ms.HasFinalityProvider(ctx, fpBTCPK) {
			return nil, types.ErrFpNotFound.Wrapf("finality provider pk: %s", fpBTCPK.MarshalHex())
		}
	}

	// Parse staking tx
	stakingMsgTx, err := bbn.NewBTCTxFromBytes(req.StakingTx.Transaction)
	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("cannot be parsed: %v", err)
	}

	// Check staking tx is not duplicated
	stakingTxHash := stakingMsgTx.TxHash()
	delgation := ms.getBTCDelegation(ctx, stakingTxHash)
	if delgation != nil {
		return nil, types.ErrReusedStakingTx.Wrapf("duplicated tx hash: %s", stakingTxHash.String())
	}

	// Check if data provided in request, matches data to which staking tx is committed
	fpPKs, err := bbn.NewBTCPKsFromBIP340PKs(req.FpBtcPkList)
	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("cannot parse finality provider PK list: %v", err)
	}
	covenantPKs, err := bbn.NewBTCPKsFromBIP340PKs(vp.Params.CovenantPks)
	if err != nil {
		// programming error
		panic("failed to parse covenant PKs in KVStore")
	}
	stakerPk := req.BtcPk.MustToBTCPK()

	stakingInfo, err := btcstaking.BuildStakingInfo(
		stakerPk,
		fpPKs,
		covenantPKs,
		vp.Params.CovenantQuorum,
		uint16(req.StakingTime),
		btcutil.Amount(req.StakingValue),
		ms.btcNet,
	)
	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("err: %v", err)
	}

	stakingOutputIdx, err := bbn.GetOutputIdxInBTCTx(stakingMsgTx, stakingInfo.StakingOutput)
	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrap("staking tx does not contain expected staking output")
	}

	// Check staking tx timelock has correct values
	// get startheight and endheight of the timelock
	stakingTxHeader := ms.btclcKeeper.GetHeaderByHash(ctx, req.StakingTx.Key.Hash)
	if stakingTxHeader == nil {
		return nil, fmt.Errorf("header that includes the staking tx is not found")
	}
	startHeight := stakingTxHeader.Height
	endHeight := stakingTxHeader.Height + uint64(req.StakingTime)

	// ensure staking tx is k-deep
	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	stakingTxDepth := btcTip.Height - stakingTxHeader.Height
	if stakingTxDepth < kValue {
		return nil, types.ErrInvalidStakingTx.Wrapf("not k-deep: k=%d; depth=%d", kValue, stakingTxDepth)
	}
	// ensure staking tx's timelock has more than w BTC blocks left
	if btcTip.Height+wValue >= endHeight {
		return nil, types.ErrInvalidStakingTx.Wrapf("staking tx's timelock has no more than w(=%d) blocks left", wValue)
	}

	// verify staking tx info, i.e., inclusion proof
	if err := req.StakingTx.VerifyInclusion(stakingTxHeader.Header, ms.btccKeeper.GetPowLimit()); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("not included in the Bitcoin chain: %v", err)
	}

	// check slashing tx and its consistency with staking tx
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	// decode slashing address
	// TODO: Decode slashing address only once, as it is the same for all BTC delegations
	slashingAddr, err := btcutil.DecodeAddress(vp.Params.SlashingAddress, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to decode slashing address in genesis: %w", err))
	}

	// Check slashing tx and staking tx are valid and consistent
	if err := btcstaking.CheckTransactions(
		slashingMsgTx,
		stakingMsgTx,
		stakingOutputIdx,
		vp.Params.MinSlashingTxFeeSat,
		vp.Params.SlashingRate,
		slashingAddr,
		stakerPk,
		validatedUnbondingTime,
		ms.btcNet,
	); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrap(err.Error())
	}

	// verify delegator sig against slashing path of the staking tx's script
	slashingSpendInfo, err := stakingInfo.SlashingPathSpendInfo()
	if err != nil {
		panic(fmt.Errorf("failed to construct slashing path from the staking tx: %w", err))
	}

	err = req.SlashingTx.VerifySignature(
		stakingInfo.StakingOutput.PkScript,
		stakingInfo.StakingOutput.Value,
		slashingSpendInfo.GetPkScriptPath(),
		stakerPk,
		req.DelegatorSlashingSig,
	)
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("invalid delegator signature: %v", err)
	}

	// all good, construct BTCDelegation and insert BTC delegation
	// NOTE: the BTC delegation does not have voting power yet. It will
	// have voting power only when 1) its corresponding staking tx is k-deep,
	// and 2) it receives a covenant signature
	newBTCDel := &types.BTCDelegation{
		BabylonPk:        req.BabylonPk,
		BtcPk:            req.BtcPk,
		Pop:              req.Pop,
		FpBtcPkList:      req.FpBtcPkList,
		StartHeight:      startHeight,
		EndHeight:        endHeight,
		TotalSat:         uint64(stakingInfo.StakingOutput.Value),
		StakingTx:        req.StakingTx.Transaction,
		StakingOutputIdx: stakingOutputIdx,
		SlashingTx:       req.SlashingTx,
		DelegatorSig:     req.DelegatorSlashingSig,
		UnbondingTime:    uint32(validatedUnbondingTime),
		CovenantSigs:     nil,        // NOTE: covenant signature will be submitted in a separate msg by covenant
		BtcUndelegation:  nil,        // this will be constructed in below code
		ParamsVersion:    vp.Version, // version of the params against delegations was validated
	}

	/*
		logics about early unbonding
	*/

	// deserialize provided transactions
	unbondingSlashingMsgTx, err := req.UnbondingSlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot convert unbonding slashing tx to wire.MsgTx: %v", err)
	}
	unbondingMsgTx, err := bbn.NewBTCTxFromBytes(req.UnbondingTx)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	// Check that unbonding tx input is pointing to staking tx
	if !unbondingMsgTx.TxIn[0].PreviousOutPoint.Hash.IsEqual(&stakingTxHash) {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("slashing transaction must spend staking output")
	}
	// Check that staking tx output index matches unbonding tx output index
	if unbondingMsgTx.TxIn[0].PreviousOutPoint.Index != stakingOutputIdx {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("slashing transaction input must spend staking output")
	}

	// building unbonding info
	unbondingInfo, err := btcstaking.BuildUnbondingInfo(
		newBTCDel.BtcPk.MustToBTCPK(),
		fpPKs,
		covenantPKs,
		vp.Params.CovenantQuorum,
		validatedUnbondingTime,
		btcutil.Amount(req.UnbondingValue),
		ms.btcNet,
	)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("err: %v", err)
	}

	// get unbonding output index
	unbondingOutputIdx, err := bbn.GetOutputIdxInBTCTx(unbondingMsgTx, unbondingInfo.UnbondingOutput)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding tx does not contain expected unbonding output")
	}

	// Check that slashing tx and unbonding tx are valid and consistent
	err = btcstaking.CheckTransactions(
		unbondingSlashingMsgTx,
		unbondingMsgTx,
		unbondingOutputIdx,
		vp.Params.MinSlashingTxFeeSat,
		vp.Params.SlashingRate,
		vp.Params.MustGetSlashingAddress(ms.btcNet),
		stakerPk,
		validatedUnbondingTime,
		ms.btcNet,
	)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("err: %v", err)
	}

	// Check staker signature against slashing path of the unbonding tx
	unbondingSlashingSpendInfo, err := unbondingInfo.SlashingPathSpendInfo()
	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}

	err = req.UnbondingSlashingTx.VerifySignature(
		unbondingInfo.UnbondingOutput.PkScript,
		unbondingInfo.UnbondingOutput.Value,
		unbondingSlashingSpendInfo.GetPkScriptPath(),
		newBTCDel.BtcPk.MustToBTCPK(),
		req.DelegatorUnbondingSlashingSig,
	)
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("invalid delegator signature: %v", err)
	}

	// Check unbonding tx fees against staking tx.
	// - fee is larger than 0
	// - ubonding output value is is at leat `MinUnbondingValue` percent of staking output value
	if unbondingMsgTx.TxOut[0].Value >= stakingMsgTx.TxOut[newBTCDel.StakingOutputIdx].Value {
		// Note: we do not enfore any minimum fee for unbonding tx, we only require that it is larger than 0
		// Given that unbonding tx must not be replacable and we do not allow sending it second time, it places
		// burden on staker to choose right fee.
		// Unbonding tx should not be replaceable at babylon level (and by extension on btc level), as this would
		// allow staker to spam the network with unbonding txs, which would force covenant and finality provider to send signatures.
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding tx fee must be larger that 0")
	}

	minUnbondingValue := caluculateMinimumUnbondingValue(stakingMsgTx.TxOut[stakingOutputIdx], &vp.Params)
	if btcutil.Amount(unbondingMsgTx.TxOut[0].Value) < minUnbondingValue {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding output value must be at least %s, based on staking output", minUnbondingValue)
	}

	// all good, add BTC undelegation
	newBTCDel.BtcUndelegation = &types.BTCUndelegation{
		UnbondingTx:              req.UnbondingTx,
		SlashingTx:               req.UnbondingSlashingTx,
		DelegatorSlashingSig:     req.DelegatorUnbondingSlashingSig,
		DelegatorUnbondingSig:    nil,
		CovenantSlashingSigs:     nil,
		CovenantUnbondingSigList: nil,
	}

	// add this BTC delegation, and emit corresponding events
	if err := ms.AddBTCDelegation(ctx, newBTCDel); err != nil {
		panic(fmt.Errorf("failed to add BTC delegation that has passed verification: %w", err))
	}

	return &types.MsgCreateBTCDelegationResponse{}, nil
}

func (ms msgServer) getBTCDelWithParams(
	ctx context.Context,
	stakingTxHash string) (*types.BTCDelegation, *types.Params, error) {
	btcDel, err := ms.GetBTCDelegation(ctx, stakingTxHash)
	if err != nil {
		return nil, nil, err
	}

	bsParams := ms.GetParamsByVersion(ctx, btcDel.ParamsVersion)
	if bsParams == nil {
		panic("params version in BTC delegation is not found")
	}

	return btcDel, bsParams, nil
}

// AddCovenantSig adds signatures from covenants to a BTC delegation
// TODO: refactor this handler. Now it's too convoluted
func (ms msgServer) AddCovenantSigs(goCtx context.Context, req *types.MsgAddCovenantSigs) (*types.MsgAddCovenantSigsResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricsKeyAddCovenantSigs)

	ctx := sdk.UnwrapSDKContext(goCtx)
	// basic stateless checks
	if err := req.ValidateBasic(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	btcDel, params, err := ms.getBTCDelWithParams(ctx, req.StakingTxHash)

	if err != nil {
		return nil, err
	}

	// ensure that the given covenant PK is in the parameter
	if !params.HasCovenantPK(req.Pk) {
		return nil, types.ErrInvalidCovenantPK.Wrapf("covenant pk: %s", req.Pk.MarshalHex())
	}

	if btcDel.IsSignedByCovMember(req.Pk) && btcDel.BtcUndelegation.IsSignedByCovMember(req.Pk) {
		ms.Logger(ctx).Debug("Received duplicated covenant signature", "covenant pk", req.Pk.MarshalHex())
		return &types.MsgAddCovenantSigsResponse{}, nil
	}

	if btcDel.HasCovenantQuorums(params.CovenantQuorum) {
		ms.Logger(ctx).Debug("Received covenant signature after achieving quorum", "covenant pk", req.Pk.MarshalHex())
		return &types.MsgAddCovenantSigsResponse{}, nil
	}

	// ensure BTC delegation is still pending, i.e., not expired
	btcTipHeight := ms.btclcKeeper.GetTipInfo(ctx).Height
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	status := btcDel.GetStatus(btcTipHeight, wValue, params.CovenantQuorum)
	if status != types.BTCDelegationStatus_PENDING {
		ms.Logger(ctx).Debug("Received covenant signature after the BTC delegation is already expired", "covenant pk", req.Pk.MarshalHex())
		return &types.MsgAddCovenantSigsResponse{}, nil
	}

	// Check that the number of covenant sigs and number of the
	// finality providers are matched
	if len(req.SlashingTxSigs) != len(btcDel.FpBtcPkList) {
		return nil, types.ErrInvalidCovenantSig.Wrapf(
			"number of covenant signatures: %d, number of finality providers being staked to: %d",
			len(req.SlashingTxSigs), len(btcDel.FpBtcPkList))
	}

	/*
		Verify each covenant adaptor signature over slashing tx
	*/
	stakingInfo, err := btcDel.GetStakingInfo(params, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to get staking info from a verified delegation: %w", err))
	}
	slashingSpendInfo, err := stakingInfo.SlashingPathSpendInfo()
	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}
	parsedSlashingAdaptorSignatures, err := btcDel.SlashingTx.ParseEncVerifyAdaptorSignatures(
		stakingInfo.StakingOutput,
		slashingSpendInfo,
		req.Pk,
		btcDel.FpBtcPkList,
		req.SlashingTxSigs,
	)
	if err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrapf("err: %v", err)
	}

	// Check that the number of covenant sigs and number of the
	// finality providers are matched
	if len(req.SlashingUnbondingTxSigs) != len(btcDel.FpBtcPkList) {
		return nil, types.ErrInvalidCovenantSig.Wrapf(
			"number of covenant signatures: %d, number of finality providers being staked to: %d",
			len(req.SlashingUnbondingTxSigs), len(btcDel.FpBtcPkList))
	}

	/*
		Verify Schnorr signature over unbonding tx
	*/
	unbondingMsgTx, err := bbn.NewBTCTxFromBytes(btcDel.BtcUndelegation.UnbondingTx)
	if err != nil {
		panic(fmt.Errorf("failed to parse unbonding tx from existing delegation with hash %s : %v", req.StakingTxHash, err))
	}
	unbondingSpendInfo, err := stakingInfo.UnbondingPathSpendInfo()
	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}
	if err := btcstaking.VerifyTransactionSigWithOutputData(
		unbondingMsgTx,
		stakingInfo.StakingOutput.PkScript,
		stakingInfo.StakingOutput.Value,
		unbondingSpendInfo.GetPkScriptPath(),
		req.Pk.MustToBTCPK(),
		*req.UnbondingTxSig,
	); err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrap(err.Error())
	}

	/*
		verify each adaptor signature on slashing unbonding tx
	*/
	unbondingOutput := unbondingMsgTx.TxOut[0] // unbonding tx always have only one output
	unbondingInfo, err := btcDel.GetUnbondingInfo(params, ms.btcNet)
	if err != nil {
		panic(err)
	}
	unbondingSlashingSpendInfo, err := unbondingInfo.SlashingPathSpendInfo()
	if err != nil {
		// our unbonding info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}
	parsedUnbondingSlashingAdaptorSignatures, err := btcDel.BtcUndelegation.SlashingTx.ParseEncVerifyAdaptorSignatures(
		unbondingOutput,
		unbondingSlashingSpendInfo,
		req.Pk,
		btcDel.FpBtcPkList,
		req.SlashingUnbondingTxSigs,
	)
	if err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrapf("err: %v", err)
	}

	// All is fine add received signatures to the BTC delegation and BtcUndelegation
	// and emit corresponding events
	ms.addCovenantSigsToBTCDelegation(
		ctx,
		btcDel,
		req.Pk,
		parsedSlashingAdaptorSignatures,
		req.UnbondingTxSig,
		parsedUnbondingSlashingAdaptorSignatures,
		params,
	)

	return &types.MsgAddCovenantSigsResponse{}, nil
}

// BTCUndelegate adds a signature on the unbonding tx from the BTC delegator
// this effectively proves that the BTC delegator wants to unbond and Babylon
// will consider its BTC delegation unbonded
func (ms msgServer) BTCUndelegate(goCtx context.Context, req *types.MsgBTCUndelegate) (*types.MsgBTCUndelegateResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricsKeyBTCUndelegate)

	ctx := sdk.UnwrapSDKContext(goCtx)
	// basic stateless checks
	if err := req.ValidateBasic(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	btcDel, bsParams, err := ms.getBTCDelWithParams(ctx, req.StakingTxHash)

	if err != nil {
		return nil, err
	}

	// ensure the BTC delegation with the given staking tx hash is active
	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	if btcDel.GetStatus(btcTip.Height, wValue, bsParams.CovenantQuorum) != types.BTCDelegationStatus_ACTIVE {
		return nil, types.ErrInvalidBTCUndelegateReq.Wrap("cannot unbond an inactive BTC delegation")
	}

	// verify the signature on unbonding tx from delegator
	unbondingMsgTx, err := bbn.NewBTCTxFromBytes(btcDel.BtcUndelegation.UnbondingTx)
	if err != nil {
		panic(fmt.Errorf("failed to parse unbonding tx from existing delegation with hash %s : %v", req.StakingTxHash, err))
	}
	stakingInfo, err := btcDel.GetStakingInfo(bsParams, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to get staking info from a verified delegation: %w", err))
	}
	unbondingSpendInfo, err := stakingInfo.UnbondingPathSpendInfo()
	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}
	if err := btcstaking.VerifyTransactionSigWithOutputData(
		unbondingMsgTx,
		stakingInfo.StakingOutput.PkScript,
		stakingInfo.StakingOutput.Value,
		unbondingSpendInfo.GetPkScriptPath(),
		btcDel.BtcPk.MustToBTCPK(),
		*req.UnbondingTxSig,
	); err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrap(err.Error())
	}

	// all good, add the signature to BTC delegation's undelegation
	// and set back
	ms.btcUndelegate(ctx, btcDel, req.UnbondingTxSig)

	return &types.MsgBTCUndelegateResponse{}, nil
}

// SelectiveSlashingEvidence handles the evidence that a finality provider has
// selectively slashed a BTC delegation
func (ms msgServer) SelectiveSlashingEvidence(goCtx context.Context, req *types.MsgSelectiveSlashingEvidence) (*types.MsgSelectiveSlashingEvidenceResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricsKeySelectiveSlashingEvidence)

	ctx := sdk.UnwrapSDKContext(goCtx)

	btcDel, bsParams, err := ms.getBTCDelWithParams(ctx, req.StakingTxHash)

	if err != nil {
		return nil, err
	}

	// ensure the BTC delegation is active, or its BTC undelegation receives an
	// unbonding signature from the staker
	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	covQuorum := bsParams.CovenantQuorum
	if btcDel.GetStatus(btcTip.Height, wValue, covQuorum) != types.BTCDelegationStatus_ACTIVE && !btcDel.IsUnbondedEarly() {
		return nil, types.ErrBTCDelegationNotFound.Wrap("a BTC delegation that is not active or unbonding early cannot be slashed")
	}

	// decode the finality provider's BTC SK/PK
	fpSK, fpPK := btcec.PrivKeyFromBytes(req.RecoveredFpBtcSk)
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(fpPK)

	// ensure the BTC delegation is staked to the given finality provider
	fpIdx := btcDel.GetFpIdx(fpBTCPK)
	if fpIdx == -1 {
		return nil, types.ErrFpNotFound.Wrapf("BTC delegation is not staked to the finality provider")
	}

	// ensure the finality provider exists and is not slashed
	fp, err := ms.GetFinalityProvider(ctx, fpBTCPK.MustMarshal())
	if err != nil {
		panic(types.ErrFpNotFound.Wrapf("failing to find the finality provider with BTC delegations"))
	}
	if fp.IsSlashed() {
		return nil, types.ErrFpAlreadySlashed
	}

	// at this point, the finality provider must have done selective slashing and must be
	// adversarial

	// slash the finality provider now
	if err := ms.SlashFinalityProvider(ctx, fpBTCPK.MustMarshal()); err != nil {
		panic(err) // failed to slash the finality provider, must be programming error
	}

	// emit selective slashing event
	evidence := &types.SelectiveSlashingEvidence{
		StakingTxHash:    req.StakingTxHash,
		FpBtcPk:          fpBTCPK,
		RecoveredFpBtcSk: fpSK.Serialize(),
	}
	event := &types.EventSelectiveSlashing{Evidence: evidence}
	if err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventSelectiveSlashing event: %w", err))
	}

	return &types.MsgSelectiveSlashingEvidenceResponse{}, nil
}
