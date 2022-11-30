package testckpt

import (
	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingkeeper "github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx                 sdk.Context
	App                 *app.BabylonApp
	CheckpointingKeeper *keeper.Keeper
	MsgSrvr             types.MsgServer
	QueryClient         types.QueryClient
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
		EpochingKeeper:      &epochingKeeper,
		GenAccs:             accs,
	}
}
