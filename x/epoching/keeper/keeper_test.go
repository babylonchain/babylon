package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testhelper "github.com/babylonchain/babylon/testutil/helper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

func TestParams(t *testing.T) {
	helper := testhelper.NewHelper(t)
	keeper := helper.App.EpochingKeeper
	ctx := helper.Ctx

	expParams := types.DefaultParams()

	// check that the empty keeper loads the default
	resParams := helper.App.EpochingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	// modify a params, save, and retrieve
	expParams.EpochInterval = 777

	if err := keeper.SetParams(ctx, expParams); err != nil {
		panic(err)
	}

	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}
