package btccheckpoint_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/btccheckpoint"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.BtcCheckpointKeeper.SetParams(ctx, types.DefaultParams())
	genesisState := btccheckpoint.ExportGenesis(ctx, app.BtcCheckpointKeeper)
	require.Equal(t, genesisState.Params, types.DefaultParams())
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisState := types.GenesisState{
		Params: types.Params{
			BtcConfirmationDepth:          999,
			CheckpointFinalizationTimeout: 888,
		},
	}

	btccheckpoint.InitGenesis(ctx, app.BtcCheckpointKeeper, genesisState)
	require.Equal(t, app.BtcCheckpointKeeper.GetParams(ctx).BtcConfirmationDepth, uint64(999))
	require.Equal(t, app.BtcCheckpointKeeper.GetParams(ctx).CheckpointFinalizationTimeout, uint64(888))
}
