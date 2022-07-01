package epoching_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/epoching"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.EpochingKeeper.SetParams(ctx, types.DefaultParams())
	genesisState := epoching.ExportGenesis(ctx, app.EpochingKeeper)
	require.Equal(t, genesisState.Params, types.DefaultParams())
}
