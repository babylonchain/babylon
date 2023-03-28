package btclightclient_test

import (
	"testing"

	bbn "github.com/babylonchain/babylon/types"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/nullify"
	"github.com/babylonchain/babylon/x/btclightclient"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	headerBytes := bbn.GetBaseBTCHeaderBytes()
	headerHeight := bbn.GetBaseBTCHeaderHeight()
	headerHash := headerBytes.Hash()
	headerWork := types.CalcWork(&headerBytes)
	baseHeaderInfo := types.NewBTCHeaderInfo(&headerBytes, headerHash, headerHeight, &headerWork)

	genesisState := types.GenesisState{
		BaseBtcHeader: *baseHeaderInfo,
	}

	k, ctx := keepertest.BTCLightClientKeeper(t)
	btclightclient.InitGenesis(ctx, *k, genesisState)
	got := btclightclient.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
