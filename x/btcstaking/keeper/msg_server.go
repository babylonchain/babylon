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
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	if ms.HasBTCValidator(ctx, *req.BtcPk) {
		return nil, types.ErrDuplicatedBTCVal
	}

	// all good, add this validator
	btcVal := types.BTCValidator{
		BabylonPk: req.BabylonPk,
		BtcPk:     req.BtcPk,
		Pop:       req.Pop,
	}
	ms.setBTCValidator(ctx, &btcVal)

	return &types.MsgCreateBTCValidatorResponse{}, nil
}

// CreateBTCDelegation creates a BTC delegation
func (ms msgServer) CreateBTCDelegation(goCtx context.Context, req *types.MsgCreateBTCDelegation) (*types.MsgCreateBTCDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// extract staking script from staking tx
	stakingOutputInfo, err := req.StakingTx.GetStakingOutputInfo(ms.btcNet)
	if err != nil {
		return nil, err
	}
	delBTCPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.StakerKey)
	valBTCPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.ValidatorKey)
	juryPK := bbn.NewBIP340PubKeyFromBTCPK(stakingOutputInfo.StakingScriptData.JuryKey)

	// ensure the staking tx is not duplicated
	// NOTE: it's okay that the same staker has multiple delegations
	// the situation that we need to prevent here is that every staking tx
	// can only correspond to a single BTC delegation
	// TODO: the current impl does not support multiple delegations with the same (valPK, delPK) pair
	// since a delegation is keyed by (valPK, delPK). Need to decide whether to support this
	btcDel, err := ms.GetBTCDelegation(ctx, *valBTCPK, *delBTCPK)
	if err == nil && btcDel.StakingTx.Equals(req.StakingTx) {
		return nil, fmt.Errorf("the BTC staking tx is already used")
	}

	// ensure staking tx is using correct jury PK
	paramJuryPK := ms.GetParams(ctx).JuryPk
	if !juryPK.Equals(paramJuryPK) {
		return nil, fmt.Errorf("staking tx specifies a wrong jury PK %s (expected: %s)", hex.EncodeToString(*juryPK), hex.EncodeToString(*paramJuryPK))
	}

	// ensure staking tx is k-deep
	stakingTxHeader, stakingTxDepth, err := ms.getHeaderAndDepth(ctx, req.StakingTxInfo.Key.Hash)
	if err != nil {
		return nil, err
	}
	kValue := ms.btccKeeper.GetParams(ctx).BtcConfirmationDepth
	if stakingTxDepth < kValue {
		return nil, fmt.Errorf("staking tx is not k-deep yet. k=%d, depth=%d", kValue, stakingTxDepth)
	}
	// verify staking tx info, i.e., inclusion proof
	if err := req.StakingTxInfo.VerifyInclusion(stakingTxHeader.Header, ms.btccKeeper.GetPowLimit()); err != nil {
		return nil, err
	}

	// check slashing tx and its consistency with staking tx
	slashingMsgTx, err := req.SlashingTx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	slashingAddr, err := btcutil.DecodeAddress(ms.GetParams(ctx).SlashingAddress, ms.btcNet)
	if err != nil {
		return nil, err
	}
	stakingMsgTx, err := req.StakingTx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	// TODO: parameterise slash min fee
	if _, err := btcstaking.CheckTransactions(slashingMsgTx, stakingMsgTx, 1, slashingAddr, req.StakingTx.StakingScript, ms.btcNet); err != nil {
		return nil, err
	}

	// verify delegator_sig
	err = req.SlashingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		req.StakingTx.StakingScript,
		stakingOutputInfo.StakingScriptData.StakerKey,
		req.DelegatorSig,
	)
	if err != nil {
		return nil, err
	}

	// all good, construct BTCDelegation and insert BTC delegation
	newBTCDel := &types.BTCDelegation{
		BabylonPk:    req.BabylonPk,
		BtcPk:        delBTCPK,
		Pop:          req.Pop,
		ValBtcPk:     valBTCPK,
		StartHeight:  stakingTxHeader.Height,
		EndHeight:    stakingTxHeader.Height + uint64(stakingOutputInfo.StakingScriptData.StakingTime),
		TotalSat:     uint64(stakingOutputInfo.StakingAmount),
		StakingTx:    req.StakingTx,
		SlashingTx:   req.SlashingTx,
		DelegatorSig: req.DelegatorSig,
		JurySig:      nil, // NOTE: jury signature will be submitted in a separate msg by jury
	}
	ms.setBTCDelegation(ctx, newBTCDel)

	return &types.MsgCreateBTCDelegationResponse{}, nil
}

// AddJurySig adds a signature from jury to a BTC delegation
func (ms msgServer) AddJurySig(goCtx context.Context, req *types.MsgAddJurySig) (*types.MsgAddJurySigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure BTC delegation exists
	btcDel, err := ms.GetBTCDelegation(ctx, *req.ValPk, *req.DelPk)
	if err != nil {
		return nil, err
	}
	if btcDel.IsActivated() {
		return nil, fmt.Errorf("the BTC delegation has already been signed by the jury")
	}

	stakingOutputInfo, err := btcDel.StakingTx.GetStakingOutputInfo(ms.btcNet)
	if err != nil {
		// failing to get staking output info from a verified staking tx is a programming error
		panic(fmt.Errorf("failed to get staking output info from a verified staking tx"))
	}

	juryPK, err := ms.GetParams(ctx).JuryPk.ToBTCPK()
	if err != nil {
		// failing to cast a verified jury PK a programming error
		panic(fmt.Errorf("failed to cast a verified jury public key"))
	}

	// verify signature w.r.t. jury PK and signature
	err = btcDel.SlashingTx.VerifySignature(
		stakingOutputInfo.StakingPkScript,
		int64(stakingOutputInfo.StakingAmount),
		btcDel.StakingTx.StakingScript,
		juryPK,
		req.Sig,
	)
	if err != nil {
		return nil, err
	}

	// all good, add signature to BTC delegation and set it back to KVStore
	btcDel.JurySig = req.Sig
	ms.setBTCDelegation(ctx, btcDel)

	return &types.MsgAddJurySigResponse{}, nil
}
