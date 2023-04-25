package types_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzCumulativeWork(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numa := r.Uint64()
		numb := r.Uint64()
		biga := sdk.NewUint(numa)
		bigb := sdk.NewUint(numb)

		gotSum := types.CumulativeWork(biga, bigb)

		expectedSum := sdk.NewUint(0)
		expectedSum = expectedSum.Add(biga)
		expectedSum = expectedSum.Add(bigb)

		if !expectedSum.Equal(gotSum) {
			t.Errorf("Cumulative work does not correspond to actual one")
		}
	})
}
