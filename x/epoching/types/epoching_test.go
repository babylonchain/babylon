package types_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate a random epoch
		epochNumber := uint64(r.Int63()) + 1
		curEpochInterval := r.Uint64()%100 + 2
		firstBlockHeight := r.Uint64() + 1

		e := types.Epoch{
			EpochNumber:          epochNumber,
			CurrentEpochInterval: curEpochInterval,
			FirstBlockHeight:     firstBlockHeight,
		}

		lastBlockHeight := firstBlockHeight + curEpochInterval - 1
		require.Equal(t, lastBlockHeight, e.GetLastBlockHeight())
		secondBlockheight := firstBlockHeight + 1
		require.Equal(t, secondBlockheight, e.GetSecondBlockHeight())
	})
}
