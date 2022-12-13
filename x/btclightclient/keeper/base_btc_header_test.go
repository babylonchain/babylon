package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/keeper"
	"math/rand"
	"testing"
)

func FuzzKeeperBaseBTCHeader(f *testing.F) {
	/*
		Checks:
		1. if a BTC header does not exist GetBaseBTCHeader returns nil
		2. SetBaseBTCHeader sets the base BTC header by checking the storage
		3. GetBaseBTCHeader returns the added BTC header

		Data generation:
		- Create a header. Use it as a parameter to SetBaseBTCHeader
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := keeper.BTCLightClientKeeper(t)
		retrievedHeaderInfo := blcKeeper.GetBaseBTCHeader(ctx)
		if retrievedHeaderInfo != nil {
			t.Errorf("GetBaseBTCHeader returned a header without one being set")
		}
		headerInfo1 := datagen.GenRandomBTCHeaderInfo()
		blcKeeper.SetBaseBTCHeader(ctx, *headerInfo1)
		retrievedHeaderInfo = blcKeeper.GetBaseBTCHeader(ctx)
		if retrievedHeaderInfo == nil {
			t.Fatalf("GetBaseBTCHeader returned nil when a BaseBTCHeader had been set")
		}
		if !headerInfo1.Eq(retrievedHeaderInfo) {
			t.Errorf("GetBaseBTCHeader did not set the provided BaseBTCHeader")
		}
	})
}
