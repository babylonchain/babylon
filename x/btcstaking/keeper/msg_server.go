package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
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

func mustGetSlashingAddress(params *types.Params, btcParams *chaincfg.Params) btcutil.Address {
	slashingAddr, err := btcutil.DecodeAddress(params.SlashingAddress, btcParams)
	if err != nil {
		panic(fmt.Errorf("failed to decode slashing address in genesis: %w", err))
	}
	return slashingAddr
}

func mustGetStakingTxInfo(del *types.BTCDelegation, params *chaincfg.Params) (*wire.MsgTx, uint32) {
	stakingTxMsg, err := del.StakingTx.ToMsgTx()

	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed deserialize staking tx from db"))
	}

	stakingOutputIndex, err := btcstaking.GetIdxOutputCommitingToScript(
		stakingTxMsg,
		del.StakingTx.Script,
		params,
	)

	if err != nil {
		panic(fmt.Errorf("script not matching staking tx in database"))
	}
	return stakingTxMsg, uint32(stakingOutputIndex)
}

// UpdateParams updates the params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// CreateBTCValidator creates a BTC validator
func (ms msgServer) CreateBTCValidator(goCtx context.Context, req *types.MsgCreateBTCValidator) (*types.MsgCreateBTCValidatorResponse, error) {
	// ensure the validator address does not exist before
	ctx := sdk.UnwrapSDKContext(goCtx)

	// verify proof of possession
	if err := req.Pop.Verify(req.BabylonPk, req.BtcPk, ms.btcNet); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid proof of possession: %v", err)
	}

	// ensure commission rate is at least the minimum commission rate in parameters
	if req.Commission.LT(ms.MinCommissionRate(ctx)) {
		return nil, types.ErrCommissionLTMinRate.Wrapf("cannot set validator commission to less than minimum rate of %s", ms.MinCommissionRate(ctx))
	}

	// ensure BTC validator does not exist before
	if ms.HasBTCValidator(ctx, *req.BtcPk) {
		return nil, types.ErrDuplicatedBTCVal
	}

	// all good, add this validator
	btcVal := types.BTCValidator{
		Description: req.Description,
		Commission:  req.Commission,
		BabylonPk:   req.BabylonPk,
		BtcPk:       req.BtcPk,
		Pop:         req.Pop,
	}
	ms.SetBTCValidator(ctx, &btcVal)

	// notify subscriber
	if err := ctx.EventManager().EmitTypedEvent(&types.EventNewBTCValidator{BtcVal: &btcVal}); err != nil {
		return nil, err
	}

	return &types.MsgCreateBTCValidatorResponse{}, nil
}

// CreateBTCDelegation creates a BTC delegation
func (ms msgServer) CreateBTCDelegation(goCtx context.Context, req *types.MsgCreateBTCDelegation) (*types.MsgCreateBTCDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := ms.GetParams(ctx)
	btccParams := ms.btccKeeper.GetParams(ctx)
	kValue, wValue := btccParams.BtcConfirmationDepth, btccParams.CheckpointFinalizationTimeout

	// extract staking script from staking tx
	stakingOutputInfo, err := req.StakingTx.GetBabylonOutputInfo(ms.btcNet)
	if err != nil {
		return nil, err
	}
	delBTCPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.StakerKey)
	valBTCPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.ValidatorKey)
	covenantPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.CovenantKey)

	// verify proof of possession
	if err := req.Pop.Verify(req.BabylonPk, delBTCPK, ms.btcNet); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid proof of possession: %v", err)
	}

	// extract staking tx and its hash
	stakingMsgTx, err := req.StakingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}
	stakingTxHash := stakingMsgTx.TxHash()

	// ensure the validator exists
	if !ms.HasBTCValidator(ctx, *valBTCPK) {
		return nil, types.ErrBTCValNotFound
	}

	// ensure the staking tx is not duplicated
	btcDelIndex, err := ms.getBTCDelegatorDelegationIndex(ctx, valBTCPK, delBTCPK)
	if err == nil {
		// err is nil, meaning there exists a BTC delegation for this validator and delegator
		// ensure the staking tx is not duplicated
		if btcDelIndex.Has(stakingTxHash) {
			return nil, types.ErrReusedStakingTx
		}
	}

	// ensure staking tx is using correct covenant PK
	paramCovenantPK := params.CovenantPk
	if !covenantPK.Equals(paramCovenantPK) {
		return nil, types.ErrInvalidCovenantPK.Wrapf("expected: %s; actual: %s", hex.EncodeToString(*paramCovenantPK), hex.EncodeToString(*covenantPK))
	}

	// get startheight and endheight of the timelock
	stakingTxHeader := ms.btclcKeeper.GetHeaderByHash(ctx, req.StakingTxInfo.Key.Hash)
	if stakingTxHeader == nil {
		return nil, fmt.Errorf("header that includes the staking tx is not found")
	}
	startHeight := stakingTxHeader.Height
	endHeight := stakingTxHeader.Height + uint64(stakingOutputInfo.StakingScriptData.StakingTime)

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
	if err := req.StakingTxInfo.VerifyInclusion(stakingTxHeader.Header, ms.btccKeeper.GetPowLimit()); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("not included in the Bitcoin chain: %v", err)
	}

	// check slashing tx and its consistency with staking tx
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}
	slashingAddr, err := btcutil.DecodeAddress(params.SlashingAddress, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to decode slashing address in genesis: %w", err))
	}
	if _, err := btcstaking.CheckTransactions(
		slashingMsgTx,
		stakingMsgTx,
		params.MinSlashingTxFeeSat,
		slashingAddr,
		req.StakingTx.Script,
		ms.btcNet,
	); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrap(err.Error())
	}

	// verify delegator_sig
	err = req.SlashingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		req.StakingTx.Script,
		stakingOutputInfo.StakingScriptData.StakerKey,
		req.DelegatorSig,
	)
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("invalid delegator signature: %v", err)
	}

	// all good, construct BTCDelegation and insert BTC delegation
	// NOTE: the BTC delegation does not have voting power yet. It will
	// have voting power only when 1) its corresponding staking tx is k-deep,
	// and 2) it receives a covenant signature
	newBTCDel := &types.BTCDelegation{
		BabylonPk:       req.BabylonPk,
		BtcPk:           delBTCPK,
		Pop:             req.Pop,
		ValBtcPk:        valBTCPK,
		StartHeight:     startHeight,
		EndHeight:       endHeight,
		TotalSat:        uint64(stakingOutputInfo.StakingAmount),
		StakingTx:       req.StakingTx,
		SlashingTx:      req.SlashingTx,
		DelegatorSig:    req.DelegatorSig,
		CovenantSig:     nil, // NOTE: covenant signature will be submitted in a separate msg by covenant
		BtcUndelegation: nil,
	}
	if err := ms.AddBTCDelegation(ctx, newBTCDel); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	// notify subscriber
	if err := ctx.EventManager().EmitTypedEvent(&types.EventNewBTCDelegation{BtcDel: newBTCDel}); err != nil {
		panic(fmt.Errorf("failed to emit EventNewBTCDelegation: %w", err))
	}

	return &types.MsgCreateBTCDelegationResponse{}, nil
}

// BtcUndelegate undelegates funds from existing delegation
func (ms msgServer) BTCUndelegate(goCtx context.Context, req *types.MsgBTCUndelegate) (*types.MsgBTCUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := ms.GetParams(ctx)
	slashingAddress := mustGetSlashingAddress(&params, ms.btcNet)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// 1. deserialize provided transactions
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	unbondingMsgTx, err := req.UnbondingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidUnbodningTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	// 2. basic stateless checks for unbodning tx
	if err := btcstaking.IsSimpleTransfer(unbondingMsgTx); err != nil {
		return nil, types.ErrInvalidUnbodningTx.Wrapf("invalid unbonding tx: %v", err)
	}

	// retrieve staking tx hash from unbonding tx, at this point we know that unbonding tx is a simple transfer with
	// one input and one output
	unbondingTxFundingOutpoint := unbondingMsgTx.TxIn[0].PreviousOutPoint
	stakingTxHash := unbondingTxFundingOutpoint.Hash.String()

	// 3. Check that slashing tx and unbonding tx are valid and consistent
	unbondingOutputInfo, err := btcstaking.CheckTransactions(
		slashingMsgTx,
		unbondingMsgTx,
		params.MinSlashingTxFeeSat,
		slashingAddress,
		req.UnbondingTx.Script,
		ms.btcNet,
	)

	if err != nil {
		return nil, types.ErrInvalidUnbodningTx.Wrapf("invalid unbonding tx: %v", err)
	}

	err = req.SlashingTx.VerifySignature(
		unbondingOutputInfo.StakingPkScript,
		int64(unbondingOutputInfo.StakingAmount),
		req.UnbondingTx.Script,
		unbondingOutputInfo.StakingScriptData.StakerKey,
		req.DelegatorSlashingSig,
	)
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("invalid delegator signature: %v", err)
	}

	// 4. Check unbonding time (staking time from unbonding tx) is larger than finalization time
	// Unbodning time must be strictly larger that babylon finalization time.
	if uint64(unbondingOutputInfo.StakingScriptData.StakingTime) <= wValue {
		return nil, types.ErrInvalidUnbodningTx.Wrapf("unbonding time must be larger than finalization time")
	}

	// 5. Check Covenant Key from script is consistent with params
	publicKeyInfos := types.KeyDataFromScript(unbondingOutputInfo.StakingScriptData)
	if !publicKeyInfos.CovenantKey.Equals(params.CovenantPk) {
		return nil, types.ErrInvalidCovenantPK.Wrapf(
			"expected: %s; actual: %s",
			hex.EncodeToString(*params.CovenantPk),
			hex.EncodeToString(*publicKeyInfos.CovenantKey),
		)
	}

	// 6. Check delegation exists for the given validator and delegator and given staking tx hash
	// as all keys are taken from script, it effectively check that values in delegation staking script
	// matches the values in the unbonding tx staking script
	del, err := ms.GetBTCDelegation(ctx, publicKeyInfos.ValidatorKey, publicKeyInfos.StakerKey, stakingTxHash)

	if err != nil {
		return nil, err
	}

	// 7. Check delegation state. Only active delegations can be unbonded.
	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	status := del.GetStatus(btcTip.Height, wValue)

	if status != types.BTCDelegationStatus_ACTIVE {
		return nil, types.ErrInvalidDelegationState
	}

	// 8. Check unbonding tx against staking tx.
	// - that input points to the staking tx, staking output
	// - fee is larger than 0
	stakingTxMsg, stakingOutputIndex := mustGetStakingTxInfo(del, ms.btcNet)

	// we only check index of the staking output, as we already retrieved delegation
	// by stakingTxHash computed from unbonding tx input
	if unbondingTxFundingOutpoint.Index != uint32(stakingOutputIndex) {
		return nil, types.ErrInvalidUnbodningTx.Wrapf("unbonding tx does not point to staking tx staking output")
	}

	if unbondingMsgTx.TxOut[0].Value >= stakingTxMsg.TxOut[stakingOutputIndex].Value {
		// Note: we do not enfore any minimum fee for unbonding tx, we only require that it is larger than 0
		// Given that unbonding tx must not be replacable and we do not allow sending it second time, it places
		// burden on staker to choose right fee.
		// Unbonding tx should not be replaceable at babylon level (and by extension on btc level), as this would
		// allow staker to spam the network with unbonding txs, which would force covenant and validator to send signatures.
		return nil, types.ErrInvalidUnbodningTx.Wrapf("unbonding tx fee must be larger that 0")
	}

	ud := types.BTCUndelegation{
		UnbondingTx:          req.UnbondingTx,
		SlashingTx:           req.SlashingTx,
		DelegatorSlashingSig: req.DelegatorSlashingSig,
		// following objects needs to be filled by covenant and validator
		// Jurry needs to provide two sigs:
		// - one for unbonding tx
		// - one for slashing tx of unbonding tx
		CovenantSlashingSig:   nil,
		CovenantUnbondingSig:  nil,
		ValidatorUnbondingSig: nil,
	}

	if err := ms.AddUndelegationToBTCDelegation(
		ctx,
		publicKeyInfos.ValidatorKey,
		publicKeyInfos.StakerKey,
		stakingTxHash,
		&ud); err != nil {
		panic(fmt.Errorf("failed to set BTC delegation that has passed verification: %w", err))
	}

	// notify subscriber
	event := &types.EventUnbondingBTCDelegation{
		BtcPk:           del.BtcPk,
		ValBtcPk:        del.ValBtcPk,
		StakingTxHash:   stakingTxHash,
		UnbondingTxHash: unbondingMsgTx.TxHash().String(),
	}
	if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventUnbondingBTCDelegation: %w", err))
	}

	return &types.MsgBTCUndelegateResponse{}, nil
}

// AddCovenantSig adds a signature from covenant to a BTC delegation
func (ms msgServer) AddCovenantSig(goCtx context.Context, req *types.MsgAddCovenantSig) (*types.MsgAddCovenantSigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure BTC delegation exists
	btcDel, err := ms.GetBTCDelegation(ctx, req.ValPk, req.DelPk, req.StakingTxHash)
	if err != nil {
		return nil, err
	}
	if btcDel.HasCovenantSig() {
		return nil, types.ErrDuplicatedCovenantSig
	}

	stakingOutputInfo, err := btcDel.StakingTx.GetBabylonOutputInfo(ms.btcNet)
	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed to get staking output info from a verified staking tx"))
	}

	covenantPK, err := ms.GetParams(ctx).CovenantPk.ToBTCPK()
	if err != nil {
		// failing to cast a verified covenant PK a programming error
		panic(fmt.Errorf("failed to cast a verified covenant public key"))
	}

	// verify signature w.r.t. covenant PK and signature
	err = btcDel.SlashingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		btcDel.StakingTx.Script,
		covenantPK,
		req.Sig,
	)
	if err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrap(err.Error())
	}

	// all good, add signature to BTC delegation and set it back to KVStore
	if err := ms.AddCovenantSigToBTCDelegation(ctx, req.ValPk, req.DelPk, req.StakingTxHash, req.Sig); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	// notify subscriber
	if err := ctx.EventManager().EmitTypedEvent(&types.EventActivateBTCDelegation{BtcDel: btcDel}); err != nil {
		panic(fmt.Errorf("failed to emit EventActivateBTCDelegation: %w", err))
	}

	return &types.MsgAddCovenantSigResponse{}, nil
}

func (ms msgServer) AddCovenantUnbondingSigs(
	goCtx context.Context,
	req *types.MsgAddCovenantUnbondingSigs) (*types.MsgAddCovenantUnbondingSigsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// 1. Check that delegation even exists for provided params
	btcDel, err := ms.GetBTCDelegation(ctx, req.ValPk, req.DelPk, req.StakingTxHash)
	if err != nil {
		return nil, err
	}

	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	status := btcDel.GetStatus(btcTip.Height, wValue)

	// 2. Check that we are in proper status
	if status != types.BTCDelegationStatus_UNBONDING {
		return nil, types.ErrInvalidDelegationState.Wrapf("Expected status: %s, actual: %s", types.BTCDelegationStatus_UNBONDING.String(), status.String())
	}

	// 3. Check that we did not recevie covenant signature yet
	if btcDel.BtcUndelegation.HasCovenantSigs() {
		return nil, types.ErrDuplicatedCovenantSig.Wrap("Covenant signature for undelegation already received")
	}

	// 4. Check that we already received validator signature
	if !btcDel.BtcUndelegation.HasValidatorSig() {
		// Covenant should provide signature only after validator to avoid validator and staker
		// collusion i.e sending unbonding tx to btc without leaving validator signature on babylon chain.
		// TODO: Maybe it is worth accepting signatures and just emmiting some kind of warning event ? as if this msg
		// processing fails, it will still be included on babylon chain, so anybody could still retrieve
		// all covenant signatures included in msg. And with warning we will at least have some kind of
		// indication that something is wrong.
		return nil, types.ErrUnbondingUnexpectedValidatorSig
	}

	// 4. Verify signature of unbodning tx against staking tx output
	stakingOutputInfo, err := btcDel.StakingTx.GetBabylonOutputInfo(ms.btcNet)
	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed to get staking output info from a verified staking tx"))
	}

	covenantPK, err := ms.GetParams(ctx).CovenantPk.ToBTCPK()
	if err != nil {
		// failing to cast a verified covenant PK is a programming error
		panic(fmt.Errorf("failed to cast a verified covenant public key"))
	}

	// UnbondingTx has exactly one input and one output so we may re-use the same
	// machinery as for slashing tx to verify signature
	err = btcDel.BtcUndelegation.UnbondingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		btcDel.StakingTx.Script,
		covenantPK,
		req.UnbondingTxSig,
	)
	if err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrap(err.Error())
	}

	// 5. Verify signature of slashing tx against unbonding tx output
	unbondingOutputInfo, err := btcDel.BtcUndelegation.UnbondingTx.GetBabylonOutputInfo(ms.btcNet)
	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed to get unbonding output info from a verified staking tx"))
	}

	err = btcDel.BtcUndelegation.SlashingTx.VerifySignature(
		unbondingOutputInfo.StakingPkScript,
		int64(unbondingOutputInfo.StakingAmount),
		btcDel.BtcUndelegation.UnbondingTx.Script,
		covenantPK,
		req.SlashingUnbondingTxSig,
	)
	if err != nil {
		return nil, types.ErrUnbodningInvalidValidatorSig.Wrap(err.Error())
	}

	// all good, add signature to BTC delegation and set it back to KVStore
	if err := ms.AddCovenantSigsToUndelegation(
		ctx,
		req.ValPk,
		req.DelPk,
		req.StakingTxHash,
		req.UnbondingTxSig,
		req.SlashingUnbondingTxSig); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	// if the BTC undelegation has validator sig, then after above operations the
	// BTC delegation will become unbonded
	if btcDel.BtcUndelegation.HasValidatorSig() {
		event := &types.EventUnbondedBTCDelegation{
			BtcPk:           btcDel.BtcPk,
			ValBtcPk:        btcDel.ValBtcPk,
			StakingTxHash:   req.StakingTxHash,
			UnbondingTxHash: btcDel.BtcUndelegation.UnbondingTx.MustGetTxHashStr(),
			FromState:       types.BTCDelegationStatus_UNBONDING,
		}
		if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
			panic(fmt.Errorf("failed to emit EventUnbondedBTCDelegation: %w", err))
		}
	}

	return nil, nil
}

func (ms msgServer) AddValidatorUnbondingSig(
	goCtx context.Context,
	req *types.MsgAddValidatorUnbondingSig) (*types.MsgAddValidatorUnbondingSigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// 1. Check that delegation even exists for provided params
	btcDel, err := ms.GetBTCDelegation(ctx, req.ValPk, req.DelPk, req.StakingTxHash)
	if err != nil {
		return nil, err
	}

	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	status := btcDel.GetStatus(btcTip.Height, wValue)

	// 2. Check that we are in proper status
	if status != types.BTCDelegationStatus_UNBONDING {
		return nil, types.ErrInvalidDelegationState.Wrapf("Expected status: %s, actual: %s", types.BTCDelegationStatus_UNBONDING.String(), status.String())
	}

	// 3. Check that we did not recevie validator signature yet
	if btcDel.BtcUndelegation.HasValidatorSig() {
		return nil, types.ErrUnbondingDuplicatedValidatorSig
	}

	// 4. Verify signature of unbonding tx against staking tx output
	stakingOutputInfo, err := btcDel.StakingTx.GetBabylonOutputInfo(ms.btcNet)
	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed to get staking output info from a verified staking tx"))
	}

	validatorPK, err := req.ValPk.ToBTCPK()

	if err != nil {
		panic(fmt.Errorf("failed to cast a verified validator public key"))
	}

	// UnbondingTx has exactly one input and one output so we may re-use the same
	// machinery as for slashing tx to verify signature
	err = btcDel.BtcUndelegation.UnbondingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		btcDel.StakingTx.Script,
		validatorPK,
		req.UnbondingTxSig,
	)
	if err != nil {
		return nil, types.ErrUnbodningInvalidValidatorSig.Wrap(err.Error())
	}

	// all good, add signature to BTC delegation and set it back to KVStore
	if err := ms.AddValidatorSigToUndelegation(ctx, req.ValPk, req.DelPk, req.StakingTxHash, req.UnbondingTxSig); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	// if the BTC undelegation has covenant sigs, then after above operations the
	// BTC delegation will become unbonded
	if btcDel.BtcUndelegation.HasCovenantSigs() {
		event := &types.EventUnbondedBTCDelegation{
			BtcPk:           btcDel.BtcPk,
			ValBtcPk:        btcDel.ValBtcPk,
			StakingTxHash:   req.StakingTxHash,
			UnbondingTxHash: btcDel.BtcUndelegation.UnbondingTx.MustGetTxHashStr(),
			FromState:       types.BTCDelegationStatus_UNBONDING,
		}
		if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
			panic(fmt.Errorf("failed to emit EventUnbondedBTCDelegation: %w", err))
		}
	}

	return nil, nil
}
