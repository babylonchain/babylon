package monitor_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/monitor"
	"github.com/stretchr/testify/require"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/monitor/types"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false)
	genesisState := monitor.ExportGenesis(ctx, app.MonitorKeeper)
	require.Equal(t, genesisState, types.DefaultGenesis())
}
