package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzRewardGaugeQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		keeper, ctx := testkeeper.IncentiveKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		// generate a list of random RewardGauges and insert them to KVStore
		rgList := []*types.RewardGauge{}
		sTypeList := []types.StakeholderType{}
		sAddrList := []sdk.AccAddress{}
		numRgs := datagen.RandomInt(r, 100)
		for i := uint64(0); i < numRgs; i++ {
			sType := datagen.GenRandomStakeholderType(r)
			sTypeList = append(sTypeList, sType)
			sAddr := datagen.GenRandomAccount().GetAddress()
			sAddrList = append(sAddrList, sAddr)
			rg := datagen.GenRandomRewardGauge(r)
			rgList = append(rgList, rg)

			keeper.SetRewardGauge(ctx, sType, sAddr, rg)
		}

		// query existence and assert consistency
		for i := range rgList {
			req := &types.QueryRewardGaugeRequest{
				Type:    sTypeList[i].String(),
				Address: sAddrList[i].String(),
			}
			resp, err := keeper.RewardGauge(wctx, req)
			require.NoError(t, err)
			require.True(t, resp.RewardGauge.Coins.IsEqual(rgList[i].Coins))
		}
	})
}
