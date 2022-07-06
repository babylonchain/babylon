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

type KeeperTestSuite struct {
	suite.Suite

	app         *app.BabylonApp
	ctx         sdk.Context
	keeper      *keeper.Keeper
	msgSrvr     types.MsgServer
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	suite.app, suite.ctx, suite.keeper, suite.msgSrvr, suite.queryClient = app, ctx, &epochingKeeper, msgSrvr, queryClient
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
