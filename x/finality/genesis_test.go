package finality_test

import (
	"testing"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/nullify"
	"github.com/babylonchain/babylon/x/finality"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.FinalityKeeper(t, nil)
	finality.InitGenesis(ctx, *k, genesisState)
	got := finality.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

}
