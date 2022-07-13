package keeper_test

import (
	"math/rand"
	"testing"
)

func FuzzKeeperBaseBTCHeader(f *testing.F) {
	/*
		Checks:
		1. if a BTC header does not exist GetBaseBTCHeader returns nil
		2. SetBaseBTCHeader sets the base BTC header by checking the storage
		3. GetBaseBTCHeader returns the added BTC header
		4. SetBaseBTCHeader fails if a BTC header has been set

		Data generation:
		- Create two headers. Use them as parameters to SetBaseBTCHeader
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}
