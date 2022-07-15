package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func setupTestKeeperWithValSet(t *testing.T) (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient, *tmtypes.ValidatorSet) {
	// generate the validator set with 10 validators
	valSet, err := testepoching.GenTmValidatorSet(10)
	require.NoError(t, err)

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

// setupTestKeeper creates a simulated Babylon app
func setupTestKeeper(t *testing.T) (*app.BabylonApp, sdk.Context, *keeper.Keeper, types.MsgServer, types.QueryClient, *tmtypes.ValidatorSet) {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	valSet := app.StakingKeeper.GetLastValidators(ctx)
	tmValSet, err := testepoching.ToTmValidators(valSet, sdk.DefaultPowerReduction)
	require.NoError(t, err)

	return app, ctx, &epochingKeeper, msgSrvr, queryClient, tmtypes.NewValidatorSet(tmValSet)
}

func genAndApplyEmptyBlock(app *app.BabylonApp, ctx sdk.Context) sdk.Context {
	newHeight := app.LastBlockHeight() + 1
	valSet := app.StakingKeeper.GetLastValidators(ctx)
	valhash := testepoching.CalculateValHash(valSet)
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
