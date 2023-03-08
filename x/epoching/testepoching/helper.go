package testepoching

import (
	"testing"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	appparams "github.com/babylonchain/babylon/app/params"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

type ValidatorInfo struct {
	BlsKey  bls12381.PrivateKey
	Address sdk.ValAddress
}

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx            sdk.Context
	App            *app.BabylonApp
	EpochingKeeper *keeper.Keeper
	MsgSrvr        types.MsgServer
	QueryClient    types.QueryClient
	StakingKeeper  *stakingkeeper.Keeper

	GenAccs        []authtypes.GenesisAccount
	ValBlsPrivKeys []ValidatorInfo
}

// NewHelper creates the helper for testing the epoching module
func NewHelper(t *testing.T) *Helper {
	app := app.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper

	// add BLS pubkey to the genesis validator
	valSet := epochingKeeper.GetValidatorSet(ctx, 0)
	require.Len(t, valSet, 1)
	genesisVal := valSet[0]
	blsPrivKey := bls12381.GenPrivKey()
	genesisBLSPubkey := blsPrivKey.PubKey()
	err := app.CheckpointingKeeper.CreateRegistration(ctx, genesisBLSPubkey, genesisVal.Addr)
	require.NoError(t, err)

	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{
		t,
		ctx,
		app,
		&epochingKeeper,
		msgSrvr,
		queryClient,
		app.StakingKeeper,
		nil,
		[]ValidatorInfo{ValidatorInfo{
			blsPrivKey,
			genesisVal.Addr,
		}},
	}
}

// NewHelperWithValSet is same as NewHelper, except that it creates a set of validators
func NewHelperWithValSet(t *testing.T) *Helper {
	// generate the validator set with 10 validators
	tmValSet, err := GenTmValidatorSet(10)
	require.NoError(t, err)

	// generate the genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	// ensure the genesis account has a sufficient amount of tokens
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdk.DefaultPowerReduction.MulRaw(10000000))),
	}
	GenAccs := []authtypes.GenesisAccount{acc}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, tmValSet, GenAccs, balance)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// get necessary subsets of the app/keeper
	epochingKeeper := app.EpochingKeeper
	valInfos := []ValidatorInfo{}
	// add BLS pubkey to the genesis validator
	valSet := epochingKeeper.GetValidatorSet(ctx, 0)
	for _, val := range valSet {
		blsPrivKey := bls12381.GenPrivKey()
		valInfos = append(valInfos, ValidatorInfo{blsPrivKey, val.Addr})
		blsPubkey := blsPrivKey.PubKey()
		err = app.CheckpointingKeeper.CreateRegistration(ctx, blsPubkey, val.Addr)
		require.NoError(t, err)
	}
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{t, ctx, app, &epochingKeeper, msgSrvr, queryClient, app.StakingKeeper, GenAccs, valInfos}
}

// GenAndApplyEmptyBlock generates a new empty block and appends it to the current blockchain
func (h *Helper) GenAndApplyEmptyBlock() sdk.Context {
	newHeight := h.App.LastBlockHeight() + 1
	valSet := h.StakingKeeper.GetLastValidators(h.Ctx)
	valhash := CalculateValHash(valSet)
	newHeader := tmproto.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
		AppHash:            datagen.GenRandomByteArray(32),
		LastCommitHash:     datagen.GenRandomLastCommitHash(),
	}

	h.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	h.App.EndBlock(abci.RequestEndBlock{})
	h.App.Commit()

	h.Ctx = h.Ctx.WithBlockHeader(newHeader)
	return h.Ctx
}

func (h *Helper) BeginBlock() sdk.Context {
	newHeight := h.App.LastBlockHeight() + 1
	valSet := h.StakingKeeper.GetLastValidators(h.Ctx)
	valhash := CalculateValHash(valSet)
	newHeader := tmproto.Header{
		Height:             newHeight,
		AppHash:            datagen.GenRandomByteArray(32),
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
	}

	h.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	h.Ctx = h.Ctx.WithBlockHeader(newHeader)
	return h.Ctx
}

func (h *Helper) EndBlock() sdk.Context {
	h.App.EndBlock(abci.RequestEndBlock{})
	h.App.Commit()
	return h.Ctx
}

// WrappedDelegate calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount math.Int) *sdk.Result {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedDelegateWithPower calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) *sdk.Result {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, h.StakingKeeper.TokensFromConsensusPower(h.Ctx, power))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedUndelegate calls handler to unbound some stake from a validator.
func (h *Helper) WrappedUndelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount math.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	wmsg := types.NewMsgWrappedUndelegate(msg)
	return h.Handle(wmsg, true)
}

// WrappedBeginRedelegate calls handler to redelegate some stake from a validator to another
func (h *Helper) WrappedBeginRedelegate(delegator sdk.AccAddress, srcVal sdk.ValAddress, dstVal sdk.ValAddress, amount math.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(appparams.DefaultBondDenom, amount)
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
