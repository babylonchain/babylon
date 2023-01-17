package monitor_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/monitor"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
