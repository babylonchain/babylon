package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	btclckeeper "github.com/babylonchain/babylon/x/btclightclient/keeper"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func allFieldsEqual(a *btclctypes.BTCHeaderInfo, b *btclctypes.BTCHeaderInfo) bool {
	return a.Height == b.Height && a.Hash.Eq(b.Hash) && a.Header.Eq(b.Header) && a.Work.Equal(*b.Work)
}

// this function must not be used at difficulty adjustment boundaries, as then
// difficulty adjustment calculation will fail
func genRandomChain(
	t *testing.T,
	r *rand.Rand,
	k *btclckeeper.Keeper,
	ctx context.Context,
	initialHeight uint64,
	chainLength uint64,
) *datagen.BTCHeaderPartialChain {
	initHeader := k.GetHeaderByHeight(ctx, initialHeight)
	randomChain := datagen.NewBTCHeaderChainFromParentInfo(
		r,
		initHeader,
		uint32(chainLength),
	)
	err := k.InsertHeaders(ctx, randomChain.ChainToBytes())
	require.NoError(t, err)
	tip := k.GetTipInfo(ctx)
	randomChainTipInfo := randomChain.GetTipInfo()
	require.True(t, allFieldsEqual(tip, randomChainTipInfo))
	return randomChain
}

func FuzzGetHeadersToBroadcast(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		babylonApp := app.Setup(t, false)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		btclcKeeper := babylonApp.BTCLightClientKeeper
		ctx := babylonApp.NewContext(false)

		hooks := zcKeeper.Hooks()

		// insert a random number of BTC headers to BTC light client
		wValue := babylonApp.BtcCheckpointKeeper.GetParams(ctx).CheckpointFinalizationTimeout
		chainLength := datagen.RandomInt(r, 10) + wValue
		genRandomChain(
			t,
			r,
			&btclcKeeper,
			ctx,
			0,
			chainLength,
		)

		// finalise a random epoch
		epochNum := datagen.RandomInt(r, 10)
		err := hooks.AfterRawCheckpointFinalized(ctx, epochNum)
		require.NoError(t, err)
		// current tip
		btcTip := btclcKeeper.GetTipInfo(ctx)
		// assert the last segment is the last w+1 BTC headers
		lastSegment := zcKeeper.GetLastSentSegment(ctx)
		require.Len(t, lastSegment.BtcHeaders, int(wValue)+1)
		for i := range lastSegment.BtcHeaders {
			require.Equal(t, btclcKeeper.GetHeaderByHeight(ctx, btcTip.Height-wValue+uint64(i)), lastSegment.BtcHeaders[i])
		}

		// finalise another epoch, during which a small number of new BTC headers are inserted
		epochNum += 1
		chainLength = datagen.RandomInt(r, 10) + 1
		genRandomChain(
			t,
			r,
			&btclcKeeper,
			ctx,
			btcTip.Height,
			chainLength,
		)
		err = hooks.AfterRawCheckpointFinalized(ctx, epochNum)
		require.NoError(t, err)
		// assert the last segment is since the header after the last tip
		lastSegment = zcKeeper.GetLastSentSegment(ctx)
		require.Len(t, lastSegment.BtcHeaders, int(chainLength))
		for i := range lastSegment.BtcHeaders {
			require.Equal(t, btclcKeeper.GetHeaderByHeight(ctx, uint64(i)+btcTip.Height+1), lastSegment.BtcHeaders[i])
		}

		// remember the current tip and the segment length
		btcTip = btclcKeeper.GetTipInfo(ctx)
		lastSegmentLength := uint64(len(lastSegment.BtcHeaders))

		// finalise another epoch, during which a number of new BTC headers with reorg are inserted
		epochNum += 1
		// reorg at a super random point
		// NOTE: it's possible that the last segment is totally reverted. We want to be resilient against
		// this, by sending the BTC headers since the last reorg point
		reorgPoint := datagen.RandomInt(r, int(btcTip.Height))
		revertedChainLength := btcTip.Height - reorgPoint
		// the fork chain needs to be longer than the canonical one
		forkChainLength := revertedChainLength + datagen.RandomInt(r, 10) + 1
		genRandomChain(
			t,
			r,
			&btclcKeeper,
			ctx,
			reorgPoint,
			forkChainLength,
		)
		err = hooks.AfterRawCheckpointFinalized(ctx, epochNum)
		require.NoError(t, err)
		// current tip
		btcTip = btclcKeeper.GetTipInfo(ctx)
		// assert the last segment is the last w+1 BTC headers
		lastSegment = zcKeeper.GetLastSentSegment(ctx)
		if revertedChainLength >= lastSegmentLength {
			// the entire last segment is reverted, the last w+1 BTC headers should be sent
			require.Len(t, lastSegment.BtcHeaders, int(wValue)+1)
			// assert the consistency of w+1 sent BTC headers
			for i := range lastSegment.BtcHeaders {
				expectedHeight := btcTip.Height - wValue + uint64(i)
				require.Equal(t, btclcKeeper.GetHeaderByHeight(ctx, expectedHeight), lastSegment.BtcHeaders[i])
			}
		} else {
			// only a subset headers of last segment are reverted, only the new fork should be sent
			require.Len(t, lastSegment.BtcHeaders, int(forkChainLength))
			// assert the consistency of the sent fork BTC headers
			for i := range lastSegment.BtcHeaders {
				expectedHeight := btcTip.Height - forkChainLength + 1 + uint64(i)
				require.Equal(t, btclcKeeper.GetHeaderByHeight(ctx, expectedHeight), lastSegment.BtcHeaders[i])
			}
		}
	})
}
