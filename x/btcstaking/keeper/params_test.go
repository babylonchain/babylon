package keeper_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestGetParamsVersions(t *testing.T) {
	k, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
	params := types.DefaultParams()

	pv := k.GetParamsWithVersion(ctx)

	require.EqualValues(t, params, pv.Params)
	require.EqualValues(t, uint32(0), pv.Version)

	params1 := types.DefaultParams()
	params1.MinSlashingTxFeeSat = 23400

	err := k.SetParams(ctx, params1)
	require.NoError(t, err)

	pv = k.GetParamsWithVersion(ctx)
	p := k.GetParams(ctx)
	require.EqualValues(t, params1, pv.Params)
	require.EqualValues(t, params1, p)
	require.EqualValues(t, uint32(1), pv.Version)

	pv0 := k.GetParamsByVersion(ctx, 0)
	require.NotNil(t, pv0)
	require.EqualValues(t, params, *pv0)
	pv1 := k.GetParamsByVersion(ctx, 1)
	require.NotNil(t, pv1)
	require.EqualValues(t, params1, *pv1)
}

// Property: All public methods related to params are consistent with each other
func FuzzParamsVersioning(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		k, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		numVersionsToGenerate := r.Intn(100) + 1
		params0 := k.GetParams(ctx)
		var generatedParams []*types.Params
		generatedParams = append(generatedParams, &params0)

		for i := 0; i < numVersionsToGenerate; i++ {
			params := types.DefaultParams()
			// randomize two parameters so each params are slightly different
			params.MinSlashingTxFeeSat = r.Int63()
			params.MinUnbondingTime = uint32(r.Intn(math.MaxUint16))
			err := k.SetParams(ctx, params)
			require.NoError(t, err)
			generatedParams = append(generatedParams, &params)
		}

		allParams := k.GetAllParams(ctx)

		require.Equal(t, len(generatedParams), len(allParams))

		for i := 0; i < len(generatedParams); i++ {
			// Check that params from aggregate query are ok
			require.EqualValues(t, *generatedParams[i], *allParams[i])

			// Check retrieval by version is ok
			paramByVersion := k.GetParamsByVersion(ctx, uint32(i))
			require.NotNil(t, paramByVersion)
			require.EqualValues(t, *generatedParams[i], *paramByVersion)
		}

		lastParams := k.GetParams(ctx)
		lastVer := k.GetParamsByVersion(ctx, uint32(len(generatedParams)-1))
		require.EqualValues(t, *generatedParams[len(generatedParams)-1], lastParams)
		require.EqualValues(t, lastParams, *lastVer)
	})
}
