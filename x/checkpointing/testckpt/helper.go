package testckpt

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching"
	epochingkeeper "github.com/babylonchain/babylon/x/epoching/keeper"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx                 sdk.Context
	App                 *app.BabylonApp
	CheckpointingKeeper *keeper.Keeper
	MsgSrvr             types.MsgServer
	QueryClient         types.QueryClient
	StakingKeeper       *stakingkeeper.Keeper
	EpochingKeeper      *epochingkeeper.Keeper

	GenAccs []authtypes.GenesisAccount
}

// NewHelper creates the helper for testing the epoching module
func NewHelper(t *testing.T, n int) *Helper {
	accs, balances := datagen.GenRandomAccWithBalance(n)
	app := app.SetupWithGenesisAccounts(accs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	checkpointingKeeper := app.CheckpointingKeeper
	epochingKeeper := app.EpochingKeeper
	stakingKeeper := app.StakingKeeper
	querier := keeper.Querier{Keeper: checkpointingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(checkpointingKeeper)

	return &Helper{
		t:                   t,
		Ctx:                 ctx,
		App:                 app,
		CheckpointingKeeper: &checkpointingKeeper,
		MsgSrvr:             msgSrvr,
		QueryClient:         queryClient,
		StakingKeeper:       stakingKeeper,
		EpochingKeeper:      &epochingKeeper,
		GenAccs:             accs,
	}
}

// CreateValidator calls handler to create a new staking validator
func (h *Helper) CreateValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, stakeAmount math.Int, ok bool) {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, stakeAmount)
	h.createValidator(addr, pk, blsPK, pop, coin, ok)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero commission
func (h *Helper) CreateValidatorWithValPower(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, valPower int64, ok bool) math.Int {
	amount := h.StakingKeeper.TokensFromConsensusPower(h.Ctx, valPower)
	coin := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	h.createValidator(addr, pk, blsPK, pop, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (h *Helper) CreateValidatorMsg(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, stakeAmount math.Int) *types.MsgWrappedCreateValidator {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, stakeAmount)
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, ZeroCommission(), sdk.OneInt())
	require.NoError(h.t, err)
	wmsg, err := types.NewMsgWrappedCreateValidator(msg, blsPK, pop)
	require.NoError(h.t, err)
	return wmsg
}

func (h *Helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, coin sdk.Coin, ok bool) {
	msg := h.CreateValidatorMsg(addr, pk, blsPK, pop, coin.Amount)
	h.Handle(msg, ok)
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

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}
