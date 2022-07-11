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
	headerHash := headerBytes.Hash()
	// The cumulative work for the Base BTC header is only the work
	// for that particular header. This means that it is very important
	// that no forks will happen that discard the base header because we
	// will not be able to detect those. Cumulative work will build based
	// on the sum of the work of the chain starting from the base header.
	headerWork := types.CalcWork(&headerBytes)
	baseHeaderInfo := types.NewBTCHeaderInfo(&headerBytes, headerHash, types.DefaultBaseHeaderHeight, &headerWork)

	genesisState := types.GenesisState{
		Params:        types.DefaultParams(),
		BaseBtcHeader: *baseHeaderInfo,
	}

	k, ctx := keepertest.BTCLightClientKeeper(t)
	btclightclient.InitGenesis(ctx, *k, genesisState)
	got := btclightclient.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
