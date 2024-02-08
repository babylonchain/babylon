package btccheckpoint_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/btccheckpoint"
	"github.com/stretchr/testify/require"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false)

	if err := app.BtcCheckpointKeeper.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	genesisState := btccheckpoint.ExportGenesis(ctx, app.BtcCheckpointKeeper)
	require.Equal(t, genesisState.Params, types.DefaultParams())
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false)

	genesisState := types.GenesisState{
		Params: types.Params{
			BtcConfirmationDepth:          888,
			CheckpointFinalizationTimeout: 999,
			CheckpointTag:                 types.DefaultCheckpointTag,
		},
	}

	btccheckpoint.InitGenesis(ctx, app.BtcCheckpointKeeper, genesisState)
	require.Equal(t, app.BtcCheckpointKeeper.GetParams(ctx).BtcConfirmationDepth, uint64(888))
	require.Equal(t, app.BtcCheckpointKeeper.GetParams(ctx).CheckpointFinalizationTimeout, uint64(999))
}
