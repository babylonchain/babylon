package types_test

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

func FuzzCumulativeWork(f *testing.F) {
	f.Add(uint64(17), uint64(25))
	f.Fuzz(func(t *testing.T, numa uint64, numb uint64) {
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
