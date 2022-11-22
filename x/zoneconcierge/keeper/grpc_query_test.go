package keeper_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzChainList(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, _, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times with random chain IDs
		numHeaders := datagen.RandomInt(100)
		expectedChainIDs := []string{}
		for i := uint64(0); i < numHeaders; i++ {
			var chainID string
			// simulate the scenario that some headers belong to the same chain
			if i > 0 && datagen.OneInN(2) {
				chainID = expectedChainIDs[rand.Intn(len(expectedChainIDs))]
			} else {
				chainID = datagen.GenRandomHexStr(30)
				expectedChainIDs = append(expectedChainIDs, chainID)
			}
			header := datagen.GenRandomIBCTMHeader(chainID, 0)
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// make query to get actual chain IDs
		resp, err := zcKeeper.ChainList(ctx, &zctypes.QueryChainListRequest{})
		require.NoError(t, err)
		actualChainIDs := resp.ChainIds

		// sort them and assert equality
		sort.Strings(expectedChainIDs)
		sort.Strings(actualChainIDs)
		require.Equal(t, len(expectedChainIDs), len(actualChainIDs))
		for i := 0; i < len(expectedChainIDs); i++ {
			require.Equal(t, expectedChainIDs[i], actualChainIDs[i])
		}
	})
}

func FuzzChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, numHeaders, numForkHeaders)

		// check if the chain info of is recorded or not
		resp, err := zcKeeper.ChainInfo(ctx, &zctypes.QueryChainInfoRequest{ChainId: czChain.ChainID})
		require.NoError(t, err)
		chainInfo := resp.ChainInfo
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}

func FuzzFinalizedChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		czChainID := string(datagen.GenRandomByteArray(10))

		btccKeeper := zctypes.NewMockBtcCheckpointKeeper(ctrl)
		epochingKeeper := zctypes.NewMockEpochingKeeper(ctrl)
		zcKeeper, ctx := testkeeper.ZoneConciergeKeeper(t, btccKeeper, epochingKeeper)
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		SimulateHeadersAndForksViaHook(ctx, hooks, czChainID, numHeaders, numForkHeaders)

		// simulate the scenario that a random epoch has ended and finalised
		epochNum := datagen.RandomInt(10)
		hooks.AfterEpochEnds(ctx, epochNum)
		hooks.AfterRawCheckpointFinalized(ctx, epochNum)

		// mock btcckeeper
		mockEpochData := &btcctypes.EpochData{
			Key: []*btcctypes.SubmissionKey{
				{Key: []*btcctypes.TransactionKey{}},
			},
			Status: btcctypes.Finalized,
		}
		btccKeeper.EXPECT().GetEpochData(gomock.Any(), gomock.Eq(epochNum)).Return(mockEpochData).AnyTimes()
		// mock epoching keeper
		epochingKeeper.EXPECT().GetHistoricalEpoch(gomock.Any(), gomock.Eq(epochNum)).Return(&epochingtypes.Epoch{}, nil).AnyTimes()

		// check if the chain info of this epoch is recorded or not
		resp, err := zcKeeper.FinalizedChainInfo(ctx, &zctypes.QueryFinalizedChainInfoRequest{ChainId: czChainID})
		require.NoError(t, err)
		chainInfo := resp.FinalizedChainInfo
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}
