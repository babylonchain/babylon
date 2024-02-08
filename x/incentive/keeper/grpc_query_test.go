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

func FuzzRewardGaugesQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		keeper, ctx := testkeeper.IncentiveKeeper(t, nil, nil, nil)

		// generate a list of random RewardGauge map and insert them to KVStore
		// where in each map, key is stakeholder type and address is the reward gauge
		rgMaps := []map[string]*types.RewardGauge{}
		sAddrList := []sdk.AccAddress{}
		numRgMaps := datagen.RandomInt(r, 100)
		for i := uint64(0); i < numRgMaps; i++ {
			rgMap := map[string]*types.RewardGauge{}
			sAddr := datagen.GenRandomAccount().GetAddress()
			sAddrList = append(sAddrList, sAddr)
			for i := uint64(0); i <= datagen.RandomInt(r, 4); i++ {
				sType := datagen.GenRandomStakeholderType(r)
				rg := datagen.GenRandomRewardGauge(r)
				rgMap[sType.String()] = rg

				keeper.SetRewardGauge(ctx, sType, sAddr, rg)
			}
			rgMaps = append(rgMaps, rgMap)
		}

		// query existence and assert consistency
		for i := range rgMaps {
			req := &types.QueryRewardGaugesRequest{
				Address: sAddrList[i].String(),
			}
			resp, err := keeper.RewardGauges(ctx, req)
			require.NoError(t, err)
			for sTypeStr, rg := range rgMaps[i] {
				require.Equal(t, rg.Coins, resp.RewardGauges[sTypeStr].Coins)
			}
		}
	})
}

func FuzzBTCStakingGaugeQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		keeper, ctx := testkeeper.IncentiveKeeper(t, nil, nil, nil)

		// generate a list of random Gauges at random heights, then insert them to KVStore
		heightList := []uint64{datagen.RandomInt(r, 1000) + 1}
		gaugeList := []*types.Gauge{datagen.GenRandomGauge(r)}
		keeper.SetBTCStakingGauge(ctx, heightList[0], gaugeList[0])

		numGauges := datagen.RandomInt(r, 100) + 1
		for i := uint64(1); i < numGauges; i++ {
			height := heightList[i-1] + datagen.RandomInt(r, 100) + 1
			heightList = append(heightList, height)
			gauge := datagen.GenRandomGauge(r)
			gaugeList = append(gaugeList, gauge)
			keeper.SetBTCStakingGauge(ctx, height, gauge)
		}

		// query existence and assert consistency
		for i := range gaugeList {
			req := &types.QueryBTCStakingGaugeRequest{
				Height: heightList[i],
			}
			resp, err := keeper.BTCStakingGauge(ctx, req)
			require.NoError(t, err)
			require.True(t, resp.Gauge.Coins.Equal(gaugeList[i].Coins))
		}
	})
}

func FuzzBTCTimestampingGaugeQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		keeper, ctx := testkeeper.IncentiveKeeper(t, nil, nil, nil)

		// initialise the 1st gauge
		epochList := []uint64{datagen.RandomInt(r, 1000) + 1}
		gaugeList := []*types.Gauge{datagen.GenRandomGauge(r)}
		keeper.SetBTCTimestampingGauge(ctx, epochList[0], gaugeList[0])

		// generate a list of random gauges at random heights, then insert them to KVStore
		numGauges := datagen.RandomInt(r, 100) + 1
		for i := uint64(1); i < numGauges; i++ {
			// increment a random number of epochs
			epoch := epochList[i-1] + datagen.RandomInt(r, 1000) + 1
			epochList = append(epochList, epoch)
			gauge := datagen.GenRandomGauge(r)
			gaugeList = append(gaugeList, gauge)
			keeper.SetBTCTimestampingGauge(ctx, epoch, gauge)
		}

		// query existence and assert consistency
		for i := range gaugeList {
			req := &types.QueryBTCTimestampingGaugeRequest{
				EpochNum: epochList[i],
			}
			resp, err := keeper.BTCTimestampingGauge(ctx, req)
			require.NoError(t, err)
			require.True(t, resp.Gauge.Coins.Equal(gaugeList[i].Coins),
				"epoch: %d\nresp.Gauge.Coins: %s\ngaugeList[i].Coins: %s\n",
				epochList[i],
				resp.Gauge.Coins.Sort().String(),
				gaugeList[i].Coins.Sort().String(),
			)
		}
	})
}
