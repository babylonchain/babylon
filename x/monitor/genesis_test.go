package monitor_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/monitor"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/monitor/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.MonitorKeeper.SetParams(ctx, types.DefaultParams())
	genesisState := monitor.ExportGenesis(ctx, app.MonitorKeeper)
	require.Equal(t, genesisState.Params, types.DefaultParams())
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisState := types.GenesisState{
		Params: types.Params{},
	}

	monitor.InitGenesis(ctx, app.MonitorKeeper, genesisState)
	require.Equal(t, app.MonitorKeeper.GetParams(ctx), genesisState.Params)
}
