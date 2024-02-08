package btclightclient_test

import (
	"testing"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/nullify"
	"github.com/babylonchain/babylon/x/btclightclient"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	baseHeaderInfo := types.SimnetGenesisBlock()
	genesisState := types.GenesisState{
		BaseBtcHeader: baseHeaderInfo,
	}

	k, ctx := keepertest.BTCLightClientKeeper(t)
	btclightclient.InitGenesis(ctx, *k, genesisState)
	got := btclightclient.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
