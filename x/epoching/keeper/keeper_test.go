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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func setupTestKeeperWithValSet(t *testing.T) (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient, *tmtypes.ValidatorSet) {
	// generate the validator set with 10 validators
	vals := []*tmtypes.Validator{}
	for i := 0; i < 10; i++ {
		privVal := datagen.NewPV()
		pubKey, err := privVal.GetPubKey()
		require.NoError(t, err)
		val := tmtypes.NewValidator(pubKey, 1)
		vals = append(vals, val)
	}
	valSet := tmtypes.NewValidatorSet(vals)

	// generate the genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// get necessary subsets of the app/keeper
	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return app, ctx, &epochingKeeper, msgSrvr, queryClient, valSet
}

func genAndApplyEmptyBlock(app *app.BabylonApp, ctx sdk.Context) sdk.Context {
	newHeight := app.LastBlockHeight() + 1
	valSet := app.StakingKeeper.GetLastValidators(ctx)
	valhash := calculateValHash(valSet)
	newHeader := tmproto.Header{
		Height:             newHeight,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	app.EndBlock(abci.RequestEndBlock{Height: newHeight})
	app.Commit()

	return ctx.WithBlockHeader(newHeader)
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

// calculate validator hash and new header
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/test_helpers.go#L156-L163)
func calculateValHash(valSet []stakingtypes.Validator) []byte {
	bzs := make([][]byte, len(valSet))
	for i, val := range valSet {
		consAddr, _ := val.GetConsAddr()
		bzs[i] = consAddr
	}
	return merkle.HashFromByteSlices(bzs)
}
