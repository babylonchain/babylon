package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func setupTestKeeperWithValSet(t *testing.T) (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient, *tmtypes.ValidatorSet) {
	t.Helper()

	// create validator set with single validator
	privVal := datagen.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	app := app.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return app, ctx, &epochingKeeper, msgSrvr, queryClient, valSet
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
