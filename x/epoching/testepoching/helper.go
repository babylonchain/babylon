package testepoching

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx            sdk.Context
	App            *app.BabylonApp
	EpochingKeeper *keeper.Keeper
	MsgSrvr        types.MsgServer
	QueryClient    types.QueryClient
	StakingKeeper  *stakingkeeper.Keeper

	GenAccs []authtypes.GenesisAccount
}

// NewHelper creates the helper for testing the epoching module
func NewHelper(t *testing.T) *Helper {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{t, ctx, app, &epochingKeeper, msgSrvr, queryClient, &app.StakingKeeper, nil}
}

// NewHelperWithValSet is same as NewHelper, except that it creates a set of validators
func NewHelperWithValSet(t *testing.T) *Helper {
	// generate the validator set with 10 validators
	valSet, err := GenTmValidatorSet(10)
	require.NoError(t, err)

	// generate the genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	// ensure the genesis account has a sufficient amount of tokens
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction.MulRaw(10000000))),
	}
	GenAccs := []authtypes.GenesisAccount{acc}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, valSet, GenAccs, balance)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// get necessary subsets of the app/keeper
	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{t, ctx, app, &epochingKeeper, msgSrvr, queryClient, &app.StakingKeeper, GenAccs}
}

// GenAndApplyEmptyBlock generates a new empty block and appends it to the current blockchain
func (h *Helper) GenAndApplyEmptyBlock() sdk.Context {
	newHeight := h.App.LastBlockHeight() + 1
	valSet := h.StakingKeeper.GetLastValidators(h.Ctx)
	valhash := CalculateValHash(valSet)
	newHeader := tmproto.Header{
		Height:             newHeight,
		AppHash:            h.App.LastCommitID().Hash,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
	}

	h.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	h.App.EndBlock(abci.RequestEndBlock{Height: newHeight})
	h.App.Commit()

	h.Ctx = h.Ctx.WithBlockHeader(newHeader)
	return h.Ctx
}

// CreateValidator calls handler to create a new staking validator
// TODO: change to the wrapped version in the checkpointing module (require modifying checkpointing module)
func (h *Helper) CreateValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int, ok bool) {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, stakeAmount)
	h.createValidator(addr, pk, coin, ok)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero commission
// TODO: change to the wrapped version in the checkpointing module (require modifying checkpointing module)
func (h *Helper) CreateValidatorWithValPower(addr sdk.ValAddress, pk cryptotypes.PubKey, valPower int64, ok bool) sdk.Int {
	amount := h.StakingKeeper.TokensFromConsensusPower(h.Ctx, valPower)
	coin := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	h.createValidator(addr, pk, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
// TODO: change to the wrapped version in the checkpointing module (require modifying checkpointing module)
func (h *Helper) CreateValidatorMsg(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int) *stakingtypes.MsgCreateValidator {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, stakeAmount)
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, ZeroCommission(), sdk.OneInt())
	require.NoError(h.t, err)
	return msg
}

// TODO: change to the wrapped version in the checkpointing module (require modifying checkpointing module)
func (h *Helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, coin sdk.Coin, ok bool) {
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, ZeroCommission(), sdk.OneInt())
	require.NoError(h.t, err)
	h.Handle(msg, ok)
}

// WrappedDelegate calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int) *sdk.Result {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedDelegateWithPower calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) *sdk.Result {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, h.StakingKeeper.TokensFromConsensusPower(h.Ctx, power))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedUndelegate calls handler to unbound some stake from a validator.
func (h *Helper) WrappedUndelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	wmsg := types.NewMsgWrappedUndelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedBeginRedelegate calls handler to redelegate some stake from a validator to another
func (h *Helper) WrappedBeginRedelegate(delegator sdk.AccAddress, srcVal sdk.ValAddress, dstVal sdk.ValAddress, amount sdk.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgBeginRedelegate(delegator, srcVal, dstVal, unbondAmt)
	wmsg := types.NewMsgWrappedBeginRedelegate(msg)
	return h.Handle(wmsg, true)
}

// Handle calls epoching handler on a given message
func (h *Helper) Handle(msg sdk.Msg, ok bool) *sdk.Result {
	handler := epoching.NewHandler(*h.EpochingKeeper)
	res, err := handler(h.Ctx, msg)
	if ok {
		require.NoError(h.t, err)
		require.NotNil(h.t, res)
	} else {
		require.Error(h.t, err)
		require.Nil(h.t, res)
	}
	return res
}

// CheckValidator asserts that a validor exists and has a given status (if status!="")
// and if has a right jailed flag.
func (h *Helper) CheckValidator(addr sdk.ValAddress, status stakingtypes.BondStatus, jailed bool) stakingtypes.Validator {
	v, ok := h.StakingKeeper.GetValidator(h.Ctx, addr)
	require.True(h.t, ok)
	require.Equal(h.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(h.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (h *Helper) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress, found bool) {
	_, ok := h.StakingKeeper.GetDelegation(h.Ctx, delegator, val)
	require.Equal(h.t, ok, found)
}
