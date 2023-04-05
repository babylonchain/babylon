package keeper_test

import (
	"testing"

	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.EpochingKeeper(t)
	params := types.DefaultParams()

	if err := k.SetParams(ctx, params); err != nil {
		panic(err)
	}

	require.EqualValues(t, params, k.GetParams(ctx))
}
