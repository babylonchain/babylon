package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzNewHeaderInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {

		// If either  of the hash strings is not of appropriate length
		// or not valid hex, generate a random hex randomly
		rand.Seed(seed)
		// Get the expected header bytes
		expectedHeaderBytes := datagen.GenRandomBTCHeaderInfo().Header
		expectedHeaderHashBytes := expectedHeaderBytes.Hash()
		height := datagen.GenRandomBTCHeight()
		work := sdk.NewUintFromBigInt(expectedHeaderBytes.Difficulty())

		headerInfo := types.NewBTCHeaderInfo(expectedHeaderBytes, expectedHeaderHashBytes, height, &work)
		// Check that all attributes are properly set
		if headerInfo == nil {
			t.Errorf("returned object is nil")
		}
		if headerInfo.Header == nil {
			t.Errorf("Header inside header info is nil")
		}
		if headerInfo.Work == nil {
			t.Errorf("Work inside header info is nil")
		}

		gotHeaderBytes := headerInfo.Header.MustMarshal()
		if bytes.Compare(expectedHeaderBytes.MustMarshal(), gotHeaderBytes) != 0 {
			t.Errorf("Expected header %s got %s", expectedHeaderBytes, gotHeaderBytes)
		}

		gotHashBytes := *headerInfo.Hash
		if bytes.Compare(expectedHeaderHashBytes.MustMarshal(), gotHashBytes.MustMarshal()) != 0 {
			t.Errorf("Expected header hash %s got %s", expectedHeaderHashBytes, gotHashBytes)
		}

		gotHeight := headerInfo.Height
		if gotHeight != height {
			t.Errorf("Expected height %d got height %d", height, gotHeight)
		}

		gotWork := headerInfo.Work
		if *gotWork != work {
			t.Errorf("Expected work %d got work %d", work.Uint64(), (*gotWork).Uint64())
		}
	})
}
