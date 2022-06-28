package btclightclient_test

import (
	bbl "github.com/babylonchain/babylon/types"
	"testing"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/nullify"
	"github.com/babylonchain/babylon/x/btclightclient"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	headerBytes, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)

	genesisState := types.GenesisState{
		Params:        types.DefaultParams(),
		BaseBtcHeader: types.DefaultBaseBTCHeader(headerBytes, types.DefaultBaseHeaderHeight),
	}

	k, ctx := keepertest.BTCLightClientKeeper(t)
	btclightclient.InitGenesis(ctx, *k, genesisState)
	got := btclightclient.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
