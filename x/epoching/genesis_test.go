package epoching_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/epoching"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	if err := app.EpochingKeeper.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	genesisState := epoching.ExportGenesis(ctx, app.EpochingKeeper)
	require.Equal(t, genesisState.Params, types.DefaultParams())
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisState := types.GenesisState{
		Params: types.Params{
			EpochInterval: 100,
		},
	}

	epoching.InitGenesis(ctx, app.EpochingKeeper, genesisState)
	require.Equal(t, app.EpochingKeeper.GetParams(ctx).EpochInterval, uint64(100))
}
