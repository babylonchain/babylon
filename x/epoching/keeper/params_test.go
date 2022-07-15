package keeper_test

import (
	"testing"

	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	_, ctx, keeper, _, _, _ := setupTestKeeper(t)

	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.EpochInterval = 777
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.EpochingKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
