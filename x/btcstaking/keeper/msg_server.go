package keeper

import (
	"context"
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/babylonchain/babylon/btcstaking"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
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

func mustGetStakingTxInfo(del *types.BTCDelegation, params *chaincfg.Params) (*wire.MsgTx, uint32) {
	stakingTxMsg, err := bbn.NewBTCTxFromBytes(del.StakingTx)

	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed deserialize staking tx from db"))
	}
	return stakingTxMsg, del.StakingOutputIdx
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

	// 1. verify proof of possession
	if err := req.Pop.Verify(req.BabylonPk, req.BtcPk, ms.btcNet); err != nil {
		return nil, types.ErrInvalidProofOfPossession.Wrapf("error while validating proof of posession: %v", err)
	}

	// 2. Ensure list of validator BTC PKs is not empty
	if len(req.ValBtcPkList) == 0 {
		return nil, types.ErrEmptyValidatorList
	}

	// 3. Ensure list of validator BTC PKs is not duplicated
	if types.ExistsDup(req.ValBtcPkList) {
		return nil, types.ErrDuplicatedValidator
	}

	// 4. Ensure all validators are known to Babylon
	for _, valBTCPK := range req.ValBtcPkList {
		if !ms.HasBTCValidator(ctx, valBTCPK) {
			return nil, types.ErrBTCValNotFound.Wrapf("validator pk: %s", valBTCPK.MarshalHex())
		}
	}

	// 5. Parse staking tx
	stakingMsgTx, err := bbn.NewBTCTxFromBytes(req.StakingTx.Transaction)

	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("cannot be parsed: %v", err)
	}

	// 6. Check staking tx is not duplicated
	stakingTxHash := stakingMsgTx.TxHash()

	delgation := ms.getBTCDelegation(ctx, stakingTxHash)

	if delgation != nil {
		return nil, types.ErrReusedStakingTx.Wrapf("duplicated tx hash: %s", stakingTxHash.String())
	}

	// 7. Check staking time is at most uint16
	if req.StakingTime > math.MaxUint16 {
		return nil, types.ErrInvalidStakingTx.Wrapf("invalid lock time: %d, max: %d", req.StakingTime, math.MaxUint16)
	}

	// 8. Check if data provided in request, matches data to which staking tx is comitted
	validatorKeys := make([]*btcec.PublicKey, 0, len(req.ValBtcPkList))
	for _, valBTCPK := range req.ValBtcPkList {
		validatorKeys = append(validatorKeys, valBTCPK.MustToBTCPK())
	}

	covenantKeys := make([]*btcec.PublicKey, 0, len(params.CovenantPks))
	for _, covenantPK := range params.CovenantPks {
		covenantKeys = append(covenantKeys, covenantPK.MustToBTCPK())
	}

	si, err := btcstaking.BuildStakingInfo(
		req.BtcPk.MustToBTCPK(),
		validatorKeys,
		covenantKeys,
		params.CovenantQuorum,
		uint16(req.StakingTime),
		btcutil.Amount(req.StakingValue),
		ms.btcNet,
	)

	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("err: %v", err)
	}

	stakingOutputIdx, err := bbn.GetOutputIdxInBTCTx(stakingMsgTx, si.StakingOutput)

	if err != nil {
		return nil, types.ErrInvalidStakingTx.Wrap("staking tx does not contain expected staking output")
	}

	// 9. Check staking tx timelock has correct values
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

	// 10. verify staking tx info, i.e., inclusion proof
	if err := req.StakingTx.VerifyInclusion(stakingTxHeader.Header, ms.btccKeeper.GetPowLimit()); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrapf("not included in the Bitcoin chain: %v", err)
	}

	// 11. check slashing tx and its consistency with staking tx
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	// decode slashing address
	// TODO: Decode slashing address only once, as it is the same for all BTC delegations
	slashingAddr, err := btcutil.DecodeAddress(params.SlashingAddress, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to decode slashing address in genesis: %w", err))
	}

	// 12. Check slashing tx and staking tx are valid and consistent
	if err := btcstaking.CheckTransactions(
		slashingMsgTx,
		stakingMsgTx,
		stakingOutputIdx,
		params.MinSlashingTxFeeSat,
		params.SlashingRate,
		slashingAddr,
		ms.btcNet,
	); err != nil {
		return nil, types.ErrInvalidStakingTx.Wrap(err.Error())
	}

	stakingOutput := stakingMsgTx.TxOut[stakingOutputIdx]

	// 13. verify delegator sig against slashing path of the script
	slashingPathInfo, err := si.SlashingPathSpendInfo()

	if err != nil {
		panic(fmt.Errorf("failed to construct slashing path from the staking tx: %w", err))
	}

	err = req.SlashingTx.VerifySignature(
		stakingOutput.PkScript,
		stakingOutput.Value,
		slashingPathInfo.GetPkScriptPath(),
		req.BtcPk.MustToBTCPK(),
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
		BabylonPk:        req.BabylonPk,
		BtcPk:            req.BtcPk,
		Pop:              req.Pop,
		ValBtcPkList:     req.ValBtcPkList,
		StartHeight:      startHeight,
		EndHeight:        endHeight,
		TotalSat:         uint64(stakingOutput.Value),
		StakingTx:        req.StakingTx.Transaction,
		StakingOutputIdx: stakingOutputIdx,
		SlashingTx:       req.SlashingTx,
		DelegatorSig:     req.DelegatorSig,
		CovenantSigs:     nil, // NOTE: covenant signature will be submitted in a separate msg by covenant
		BtcUndelegation:  nil,
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
	slashingAddress := params.MustGetSlashingAddress(ms.btcNet)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// 1. deserialize provided transactions
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	unbondingMsgTx, err := bbn.NewBTCTxFromBytes(req.UnbondingTx)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("cannot be converted to wire.MsgTx: %v", err)
	}

	// 2. basic stateless checks for unbonding tx
	if err := btcstaking.IsSimpleTransfer(unbondingMsgTx); err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("err: %v", err)
	}

	// 3. Check unbonding time (staking time from unbonding tx) is larger than finalization time
	// Unbonding time must be strictly larger that babylon finalization time.
	if uint64(req.UnbondingTime) <= wValue {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding time %d must be larger than finalization time %d", req.UnbondingTime, wValue)
	}

	// 4. Check unbonding time is lower than max uint16
	if uint64(req.UnbondingTime) > math.MaxUint16 {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding time %d must be lower than %d", req.UnbondingTime, math.MaxUint16)
	}

	// retrieve staking tx hash from unbonding tx, at this point we know that unbonding tx is a simple transfer with
	// one input and one output
	unbondingTxFundingOutpoint := unbondingMsgTx.TxIn[0].PreviousOutPoint
	stakingTxHash := unbondingTxFundingOutpoint.Hash.String()

	// 5. Check delegation wchich should be undelegeated exists and it is in correct state
	del, err := ms.GetBTCDelegation(ctx, stakingTxHash)

	if err != nil {
		return nil, types.ErrInvalidDelegationState.Wrapf("couldn't retrieve delegation for staking tx hash %s, err: %v", stakingTxHash, err)
	}

	// 6. Check delegation state. Only active delegations can be unbonded.
	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	status := del.GetStatus(btcTip.Height, wValue, params.CovenantQuorum)

	if status != types.BTCDelegationStatus_ACTIVE {
		return nil, types.ErrInvalidDelegationState.Wrapf("current status: %v, want: %s", status.String(), types.BTCDelegationStatus_ACTIVE.String())
	}

	// 7. Check unbonding tx commits to valid scripts
	validatorKeys := make([]*btcec.PublicKey, 0, len(del.ValBtcPkList))
	// We retrieve validator keys from the delegation, as we want to check that unbonding tx commits to the same
	// validator keys as staking tx.
	for _, valBTCPK := range del.ValBtcPkList {
		validatorKeys = append(validatorKeys, valBTCPK.MustToBTCPK())
	}

	covenantKeys := make([]*btcec.PublicKey, 0, len(params.CovenantPks))
	// as we do not rotate covenant keys, we can retrieve them from params
	for _, covenantPK := range params.CovenantPks {
		covenantKeys = append(covenantKeys, covenantPK.MustToBTCPK())
	}

	si, err := btcstaking.BuildUnbondingInfo(
		del.BtcPk.MustToBTCPK(),
		validatorKeys,
		covenantKeys,
		params.CovenantQuorum,
		uint16(req.UnbondingTime),
		btcutil.Amount(req.UnbondingValue),
		ms.btcNet,
	)

	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("err: %v", err)
	}

	unbondingOutputIdx, err := bbn.GetOutputIdxInBTCTx(unbondingMsgTx, si.UnbondingOutput)

	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding tx does not contain expected unbonding output")
	}

	// 8. Check that slashing tx and unbonding tx are valid and consistent
	err = btcstaking.CheckTransactions(
		slashingMsgTx,
		unbondingMsgTx,
		unbondingOutputIdx,
		params.MinSlashingTxFeeSat,
		params.SlashingRate,
		slashingAddress,
		ms.btcNet,
	)
	if err != nil {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("err: %v", err)
	}

	// 9. Check staker signature against slashing path of the unbonding tx
	unbondingOutput := unbondingMsgTx.TxOut[unbondingOutputIdx]

	slashingPathInfo, err := si.SlashingPathSpendInfo()

	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}

	err = req.SlashingTx.VerifySignature(
		unbondingOutput.PkScript,
		unbondingOutput.Value,
		slashingPathInfo.GetPkScriptPath(),
		del.BtcPk.MustToBTCPK(),
		req.DelegatorSlashingSig,
	)
	if err != nil {
		return nil, types.ErrInvalidSlashingTx.Wrapf("invalid delegator signature: %v", err)
	}

	// 8. Check unbonding tx against staking tx.
	// - that input points to the staking tx, staking output
	// - fee is larger than 0
	stakingTxMsg, stakingOutputIndex := mustGetStakingTxInfo(del, ms.btcNet)

	// we only check index of the staking output, as we already retrieved delegation
	// by stakingTxHash computed from unbonding tx input
	if unbondingTxFundingOutpoint.Index != uint32(stakingOutputIndex) {
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding tx does not point to staking tx staking output")
	}

	if unbondingMsgTx.TxOut[0].Value >= stakingTxMsg.TxOut[stakingOutputIndex].Value {
		// Note: we do not enfore any minimum fee for unbonding tx, we only require that it is larger than 0
		// Given that unbonding tx must not be replacable and we do not allow sending it second time, it places
		// burden on staker to choose right fee.
		// Unbonding tx should not be replaceable at babylon level (and by extension on btc level), as this would
		// allow staker to spam the network with unbonding txs, which would force covenant and validator to send signatures.
		return nil, types.ErrInvalidUnbondingTx.Wrapf("unbonding tx fee must be larger that 0")
	}

	ud := types.BTCUndelegation{
		UnbondingTx:          req.UnbondingTx,
		SlashingTx:           req.SlashingTx,
		DelegatorSlashingSig: req.DelegatorSlashingSig,
		// following objects needs to be filled by covenant and validator
		// covenant emulators need to provide two sigs:
		// - one for unbonding tx (schnorr sig)
		// - one for validator of the slashing tx of unbonding tx (adaptor sig)
		CovenantSlashingSigs:     nil,
		CovenantUnbondingSigList: nil,
		UnbondingTime:            req.UnbondingTime,
	}

	if err := ms.AddUndelegationToBTCDelegation(
		ctx,
		stakingTxHash,
		&ud); err != nil {
		panic(fmt.Errorf("failed to set BTC delegation that has passed verification: %w", err))
	}

	// notify subscriber
	event := &types.EventUnbondingBTCDelegation{
		BtcPk:           del.BtcPk,
		ValBtcPkList:    del.ValBtcPkList,
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
	covenantQuorum := ms.GetParams(ctx).CovenantQuorum

	// ensure BTC delegation exists
	btcDel, err := ms.GetBTCDelegation(ctx, req.StakingTxHash)
	if err != nil {
		return nil, err
	}

	stakingTx, stakingOutputIdx := mustGetStakingTxInfo(btcDel, ms.btcNet)
	stakingOutput := stakingTx.TxOut[stakingOutputIdx]

	// Note: we assume the order of adaptor sigs is matched to the
	// order of validators in the delegation
	// TODO ensure the order, currently, we only have one validator
	//  one covenant emulator
	numAdaptorSig := len(req.Sigs)
	numVals := len(btcDel.ValBtcPkList)
	if numAdaptorSig != numVals {
		return nil, types.ErrInvalidCovenantSig.Wrapf(
			"number of covenant signatures: %d, number of validators being staked to: %d",
			numAdaptorSig, numVals)
	}

	// ensure that the given covenant PK is in the parameter
	params := ms.GetParams(ctx)
	if !params.HasCovenantPK(req.Pk) {
		return nil, types.ErrInvalidCovenantPK.Wrapf("covenant pk: %s", req.Pk.MarshalHex())
	}

	spendInfo, err := btcDel.GetStakingInfo(&params, ms.btcNet)
	if err != nil {
		panic(fmt.Errorf("failed to get staking info from a verified delegation: %w", err))
	}

	slashingPathInfo, err := spendInfo.SlashingPathSpendInfo()
	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}

	// verify each covenant adaptor signature with the corresponding validator public key
	for i, sig := range req.Sigs {
		err := verifySlashingTxAdaptorSig(
			btcDel.SlashingTx,
			stakingOutput.PkScript,
			stakingOutput.Value,
			slashingPathInfo.GetPkScriptPath(),
			req.Pk.MustToBTCPK(),
			btcDel.ValBtcPkList[i].MustToBTCPK(),
			sig,
		)
		if err != nil {
			return nil, types.ErrInvalidCovenantSig.Wrapf("err: %v", err)
		}
	}

	// all good, add signatures to BTC delegation and set it back to KVStore
	if err := ms.AddCovenantSigsToBTCDelegation(ctx, req.StakingTxHash, req.Sigs, req.Pk, covenantQuorum); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	// notify subscriber
	if err := ctx.EventManager().EmitTypedEvent(&types.EventActivateBTCDelegation{BtcDel: btcDel}); err != nil {
		panic(fmt.Errorf("failed to emit EventActivateBTCDelegation: %w", err))
	}

	return &types.MsgAddCovenantSigResponse{}, nil
}

func verifySlashingTxAdaptorSig(
	slashingTx *types.BTCSlashingTx,
	stakingPkScript []byte,
	stakingAmount int64,
	stakingScript []byte,
	pk *btcec.PublicKey,
	valPk *btcec.PublicKey,
	sig []byte) error {
	adaptorSig, err := asig.NewAdaptorSignatureFromBytes(sig)
	if err != nil {
		return err
	}

	encKey, err := asig.NewEncryptionKeyFromBTCPK(valPk)
	if err != nil {
		return err
	}

	return slashingTx.EncVerifyAdaptorSignature(
		stakingPkScript,
		stakingAmount,
		stakingScript,
		pk,
		encKey,
		adaptorSig,
	)
}

func (ms msgServer) AddCovenantUnbondingSigs(
	goCtx context.Context,
	req *types.MsgAddCovenantUnbondingSigs,
) (*types.MsgAddCovenantUnbondingSigsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	wValue := ms.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	covenantQuorum := ms.GetParams(ctx).CovenantQuorum

	// 1. Check that delegation even exists for provided params
	btcDel, err := ms.GetBTCDelegation(ctx, req.StakingTxHash)
	if err != nil {
		return nil, err
	}

	btcTip := ms.btclcKeeper.GetTipInfo(ctx)
	status := btcDel.GetStatus(btcTip.Height, wValue, covenantQuorum)

	// 2. Check that we are in proper status
	if status != types.BTCDelegationStatus_UNBONDING {
		return nil, types.ErrInvalidDelegationState.Wrapf("Expected status: %s, actual: %s", types.BTCDelegationStatus_UNBONDING.String(), status.String())
	}

	// 3. Check that the number of covenant sigs and number of the
	// validators are matched
	// Note: we assume the order of adaptor sigs is matched to the
	// order of validators in the delegation
	// TODO ensure the order, currently, we only have one validator
	//  one covenant emulator
	numAdaptorSig := len(req.SlashingUnbondingTxSigs)
	numVals := len(btcDel.ValBtcPkList)
	if numAdaptorSig != numVals {
		return nil, types.ErrInvalidCovenantSig.Wrapf(
			"number of covenant signatures: %d, number of validators being staked to: %d",
			numAdaptorSig, numVals)
	}

	// 4. Verify signature of unbonding tx against staking tx output
	stakingTx, stakingOutputIdx := mustGetStakingTxInfo(btcDel, ms.btcNet)
	stakingOutput := stakingTx.TxOut[stakingOutputIdx]

	// ensure that the given covenant PK is in the parameter
	params := ms.GetParams(ctx)
	if !params.HasCovenantPK(req.Pk) {
		return nil, types.ErrInvalidCovenantPK
	}

	unbondingTxMsg, err := bbn.NewBTCTxFromBytes(btcDel.BtcUndelegation.UnbondingTx)

	if err != nil {
		panic(fmt.Errorf("failed to parse unbonding tx from existing delegation with hash %s : %v", req.StakingTxHash, err))
	}

	unbondingTxHash := unbondingTxMsg.TxHash().String()

	stakingOutputSpendInfo, err := btcDel.GetStakingInfo(&params, ms.btcNet)
	if err != nil {
		panic(err)
	}

	unbondingPathInfo, err := stakingOutputSpendInfo.UnbondingPathSpendInfo()

	if err != nil {
		// our staking info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}

	if err := btcstaking.VerifyTransactionSigWithOutputData(
		unbondingTxMsg,
		stakingOutput.PkScript,
		stakingOutput.Value,
		unbondingPathInfo.GetPkScriptPath(),
		req.Pk.MustToBTCPK(),
		*req.UnbondingTxSig,
	); err != nil {
		return nil, types.ErrInvalidCovenantSig.Wrap(err.Error())
	}

	// 5. Verify signature of slashing tx against unbonding tx output
	// unbonding tx always have only one output
	unbondingOutput := unbondingTxMsg.TxOut[0]
	unbondingInfo, err := btcDel.GetUnbondingInfo(&params, ms.btcNet)
	if err != nil {
		panic(err)
	}

	slashingPathInfo, err := unbondingInfo.SlashingPathSpendInfo()

	if err != nil {
		// our unbonding info was constructed by using BuildStakingInfo constructor, so if
		// this fails, it is a programming error
		panic(err)
	}

	// verify each covenant adaptor signature with the corresponding validator public key
	for i, sig := range req.SlashingUnbondingTxSigs {
		err := verifySlashingTxAdaptorSig(
			btcDel.BtcUndelegation.SlashingTx,
			unbondingOutput.PkScript,
			unbondingOutput.Value,
			slashingPathInfo.GetPkScriptPath(),
			req.Pk.MustToBTCPK(),
			btcDel.ValBtcPkList[i].MustToBTCPK(),
			sig,
		)
		if err != nil {
			return nil, types.ErrInvalidCovenantSig.Wrapf("err: %v", err)
		}
	}

	// all good, add signature to BTC undelegation and set it back to KVStore
	if err := ms.AddCovenantSigsToUndelegation(
		ctx,
		req.StakingTxHash,
		req.Pk,
		req.UnbondingTxSig,
		req.SlashingUnbondingTxSigs,
		covenantQuorum,
	); err != nil {
		panic("failed to set BTC delegation that has passed verification")
	}

	event := &types.EventUnbondedBTCDelegation{
		BtcPk:           btcDel.BtcPk,
		ValBtcPkList:    btcDel.ValBtcPkList,
		StakingTxHash:   req.StakingTxHash,
		UnbondingTxHash: unbondingTxHash,
		FromState:       types.BTCDelegationStatus_UNBONDING,
	}

	if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventUnbondedBTCDelegation: %w", err))
	}

	return nil, nil
}
