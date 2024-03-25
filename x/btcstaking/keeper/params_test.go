package keeper_test

import (
	"testing"

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
