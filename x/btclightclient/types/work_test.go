package types_test

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"math/big"
	"testing"
)

func FuzzCumulativeWork(f *testing.F) {
	f.Add(int64(17), int64(25))
	f.Fuzz(func(t *testing.T, numa int64, numb int64) {
		biga := big.NewInt(numa)
		bigb := big.NewInt(numb)

		gotSum := types.CumulativeWork(biga, bigb)

		expectedSum := new(big.Int)
		expectedSum.Add(biga, bigb)

		if expectedSum.Cmp(gotSum) != 0 {
			t.Errorf("Cumulative work does not correspond to actual one")
		}
	})
}
