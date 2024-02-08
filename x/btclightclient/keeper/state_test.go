package keeper_test

import (
	"math/rand"
	"testing"

	bbn "github.com/babylonchain/babylon/types"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func FuzzHeadersStateCreateHeader(f *testing.F) {
	/*
		 Checks:
		 1. A headerInfo provided as an argument leads to the following storage objects being created:
			 - A (height) -> headerInfo object
			 - A (headerHash) -> height object

		 Data generation:
		 - Create four headers:
			 1. The Base header. This will test whether the tip is set.
			 2. Create random  chain of of headers, and insert them into the state
			 3. All operations should be consistent with each other.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		state := blcKeeper.HeadersState(ctx)

		// operations no empty state
		require.Nil(t, state.GetTip())
		require.Nil(t, state.BaseHeader())
		require.False(t, state.TipExists())

		numForward := 0
		numBackward := 0
		state.IterateForwardHeaders(0, func(headerInfo *types.BTCHeaderInfo) bool {
			numForward++
			return false
		})
		state.IterateReverseHeaders(func(headerInfo *types.BTCHeaderInfo) bool {
			numBackward++
			return false
		})
		require.Equal(t, 0, numForward)
		require.Equal(t, 0, numBackward)

		_, err := state.GetHeaderByHeight(datagen.GenRandomBTCHeight(r))
		require.Error(t, err)
		rh := datagen.GenRandomBtcdHash(r)
		randomHash := bbn.NewBTCHeaderHashBytesFromChainhash(&rh)
		_, err = state.GetHeaderByHash(&randomHash)
		require.Error(t, err)

		// 10 to 60 headers
		chainLength := datagen.RandomInt(r, 50) + 10
		// height from 10 to 60
		initchainHeight := datagen.RandomInt(r, 50) + 10

		// populate the state with random chain
		_, chain := genRandomChain(
			t,
			r,
			blcKeeper,
			ctx,
			initchainHeight,
			chainLength,
		)

		// operations populates state
		require.NotNil(t, state.GetTip())
		require.NotNil(t, state.BaseHeader())
		require.True(t, state.TipExists())

		numForward = 0
		numBackward = 0
		state.IterateForwardHeaders(0, func(headerInfo *types.BTCHeaderInfo) bool {
			numForward++
			return false
		})
		state.IterateReverseHeaders(func(headerInfo *types.BTCHeaderInfo) bool {
			numBackward++
			return false
		})
		require.Equal(t, chainLength+1, uint64(numForward))
		require.Equal(t, chainLength+1, uint64(numBackward))

		chainInfos := chain.GetChainInfo()

		for _, info := range chainInfos {
			byHash, err := state.GetHeaderByHash(info.Hash)
			require.NoError(t, err)
			byHeight, err := state.GetHeaderByHeight(info.Height)
			require.NoError(t, err)
			require.True(t, allFieldsEqual(byHash, byHeight))
			require.True(t, allFieldsEqual(byHash, info))
		}
	})
}
