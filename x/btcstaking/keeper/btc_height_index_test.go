package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzBTCHeightIndex(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		keeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, nil)

		// randomise Babylon height and BTC height
		babylonHeight := datagen.RandomInt(r, 100)
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btcHeight := datagen.RandomInt(r, 100)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: btcHeight}).Times(1)
		keeper.IndexBTCHeight(ctx)

		// assert BTC height
		actualBtcHeight := keeper.GetBTCHeightAtBabylonHeight(ctx, babylonHeight)
		require.Equal(t, btcHeight, actualBtcHeight)
		// assert current BTC height
		curBtcHeight := keeper.GetCurrentBTCHeight(ctx)
		require.Equal(t, btcHeight, curBtcHeight)
	})
}
