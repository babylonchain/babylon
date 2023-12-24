package datagen

import (
	"math/rand"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

// getFirstBlockHeight returns the height of the first block of a given epoch and epoch interval
// NOTE: this is only a function for testing and assumes static epoch interval
func getFirstBlockHeight(epochNumber uint64, epochInterval uint64) uint64 {
	if epochNumber == 0 {
		return 0
	} else {
		return (epochNumber-1)*epochInterval + 1
	}
}

func GenRandomEpochNum(r *rand.Rand) uint64 {
	epochNum := r.Int63n(100)
	return uint64(epochNum)
}

func GenRandomEpochInterval(r *rand.Rand) uint64 {
	epochInterval := r.Int63n(10) + 2 // interval should be at least 2
	return uint64(epochInterval)
}

func GenRandomEpoch(r *rand.Rand) *epochingtypes.Epoch {
	epochNum := GenRandomEpochNum(r)
	epochInterval := GenRandomEpochInterval(r)
	firstBlockHeight := getFirstBlockHeight(epochNum, epochInterval)
	lastBlockHeader := GenRandomTMHeader(r, "test-chain", firstBlockHeight+epochInterval-1)
	epoch := epochingtypes.NewEpoch(
		epochNum,
		epochInterval,
		firstBlockHeight,
		&lastBlockHeader.Time,
	)
	sealerHeader := GenRandomTMHeader(r, "test-chain", firstBlockHeight+epochInterval+1) // 2nd block in the next epoch
	epoch.SealerBlockHash = GenRandomBlockHash(r)
	epoch.SealerAppHash = sealerHeader.AppHash
	return &epoch
}
