package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	monitorkeeper "github.com/babylonchain/babylon/x/monitor/keeper"
	"github.com/babylonchain/babylon/x/monitor/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func FuzzQueryEndedEpochBtcHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		lck := helper.App.BTCLightClientKeeper
		mk := helper.App.MonitorKeeper
		ek := helper.EpochingKeeper
		querier := monitorkeeper.Querier{Keeper: mk}
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, querier)
		queryClient := types.NewQueryClient(queryHelper)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock()
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// Insert header tree
		tree := datagen.NewBTCHeaderTree()
		root := lck.GetBaseBTCHeader(ctx)
		tree.Add(root, nil)
		tree.GenRandomBTCHeaderTree(1, 10, root, func(header *btclightclienttypes.BTCHeaderInfo) bool {
			err := lck.InsertHeader(ctx, header.Header)
			require.NoError(t, err)
			return true
		})

		// EndBlock of block 1
		ctx = helper.EndBlock()

		// go to BeginBlock of block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// query epoch 0 ended BTC light client height, should return base height
		req := types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 0,
		}
		resp, err := queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetBaseBTCHeader(ctx).Height, resp.BtcLightClientHeight)

		// query epoch 1 ended BTC light client height, should return tip height
		req = types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 1,
		}
		resp, err = queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetTipInfo(ctx).Height, resp.BtcLightClientHeight)
	})
}

func FuzzQueryEndedEpochBtcHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		lck := helper.App.BTCLightClientKeeper
		mk := helper.App.MonitorKeeper
		ek := helper.EpochingKeeper
		querier := monitorkeeper.Querier{Keeper: mk}
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, querier)
		queryClient := types.NewQueryClient(queryHelper)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock()
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// Insert header tree
		tree := datagen.NewBTCHeaderTree()
		root := lck.GetBaseBTCHeader(ctx)
		tree.Add(root, nil)
		tree.GenRandomBTCHeaderTree(1, 10, root, func(header *btclightclienttypes.BTCHeaderInfo) bool {
			err := lck.InsertHeader(ctx, header.Header)
			require.NoError(t, err)
			return true
		})

		// EndBlock of block 1
		ctx = helper.EndBlock()

		// go to BeginBlock of block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// query epoch 0 ended BTC light client height, should return base height
		req := types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 0,
		}
		resp, err := queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetBaseBTCHeader(ctx).Height, resp.BtcLightClientHeight)

		// query epoch 1 ended BTC light client height, should return tip height
		req = types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 1,
		}
		resp, err = queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetTipInfo(ctx).Height, resp.BtcLightClientHeight)
	})
}
