package datagen

import (
	"math/rand"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

// firstBlockHeight returns the height of the first block of a given epoch and epoch interval
// NOTE: this is only a function for testing and assumes static epoch interval
func firstBlockHeight(epochNumber uint64, epochInterval uint64) uint64 {
	if epochNumber == 0 {
		return 0
	} else {
		return (epochNumber-1)*epochInterval + 1
	}
}

func GenRandomEpochNum() uint64 {
	epochNum := rand.Int63n(100)
	return uint64(epochNum)
}

func GenRandomEpochInterval() uint64 {
	epochInterval := rand.Int63n(10) + 2 // interval should be at least 2
	return uint64(epochInterval)
}

func GenRandomEpoch() *epochingtypes.Epoch {
	epochNum := GenRandomEpochNum()
	epochInterval := GenRandomEpochInterval()
	firstBlockHeight := firstBlockHeight(epochNum, epochInterval)
	lastBlockHeader := GenRandomTMHeader("test-chain", firstBlockHeight+epochInterval-1)
	epoch := epochingtypes.NewEpoch(
		epochNum,
		epochInterval,
		firstBlockHeight,
		lastBlockHeader,
	)
	epoch.SealerHeader = GenRandomTMHeader("test-chain", firstBlockHeight+epochInterval+1) // 2nd block in the next epoch
	return &epoch
}
