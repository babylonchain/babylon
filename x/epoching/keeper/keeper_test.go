package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupTestKeeper creates a simulated Babylon app with a set of validators
func setupTestKeeperWithValSet(t *testing.T) (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient) {

	initBalances := sdk.NewIntFromBigInt(app.CoinOne).Mul(sdk.NewInt(20000))
	validator, genesisAccounts, balances := app.GenerateGenesisValidator(2, sdk.NewCoins(sdk.NewCoin("BBL", initBalances)))
	app := app.SetupWithGenesisValSet(t, validator, genesisAccounts, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return app, ctx, &epochingKeeper, msgSrvr, queryClient
}

// setupTestKeeper creates a simulated Babylon app
func setupTestKeeper() (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient) {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return app, ctx, &epochingKeeper, msgSrvr, queryClient
}

type KeeperTestSuite struct {
	suite.Suite

	app         *app.BabylonApp
	ctx         sdk.Context
	keeper      *keeper.Keeper
	msgSrvr     types.MsgServer
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx, suite.keeper, suite.msgSrvr, suite.queryClient = setupTestKeeper()
}

func TestParams(t *testing.T) {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := app.EpochingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.EpochInterval = 777
	app.EpochingKeeper.SetParams(ctx, expParams)
	resParams = app.EpochingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
