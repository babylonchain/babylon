package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpoch(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		// generate a random epoch
		epochNumber := rand.Uint64() + 1
		curEpochInterval := rand.Uint64()%100 + 2
		firstBlockHeight := rand.Uint64() + 1

		// genesis block case with some probability
		genesisFlag := rand.Uint64()%100 < 10
		if genesisFlag {
			epochNumber = 0
			firstBlockHeight = 0
		}

		e := types.Epoch{
			EpochNumber:          epochNumber,
			CurrentEpochInterval: curEpochInterval,
			FirstBlockHeight:     firstBlockHeight,
		}

		if genesisFlag {
			require.Equal(t, uint64(0), e.GetLastBlockHeight())
			require.Equal(t, uint64(0), e.GetSecondBlockHeight())
		} else {
			lastBlockHeight := firstBlockHeight + curEpochInterval - 1
			require.Equal(t, lastBlockHeight, e.GetLastBlockHeight())
			secondBlockheight := firstBlockHeight + 1
			require.Equal(t, secondBlockheight, e.GetSecondBlockHeight())
		}
	})
}
