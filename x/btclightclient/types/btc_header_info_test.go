package types_test

import (
	sdkmath "cosmossdk.io/math"

	"bytes"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btclightclient/types"
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
		work := sdkmath.NewUintFromBigInt(expectedHeaderBytes.Difficulty())

		headerInfo := types.NewBTCHeaderInfo(expectedHeaderBytes, expectedHeaderHashBytes, height, &work)
		// Check that all attributes are properly set
		if headerInfo == nil { //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
			t.Errorf("returned object is nil")
		}
		if headerInfo.Header == nil { //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
			t.Errorf("Header inside header info is nil")
		}
		if headerInfo.Work == nil { //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
			t.Errorf("Work inside header info is nil")
		}

		gotHeaderBytes := headerInfo.Header.MustMarshal() //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
		if !bytes.Equal(expectedHeaderBytes.MustMarshal(), gotHeaderBytes) {
			t.Errorf("Expected header %s got %s", expectedHeaderBytes, gotHeaderBytes)
		}

		gotHashBytes := *headerInfo.Hash //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
		if !bytes.Equal(expectedHeaderHashBytes.MustMarshal(), gotHashBytes.MustMarshal()) {
			t.Errorf("Expected header hash %s got %s", expectedHeaderHashBytes, gotHashBytes)
		}

		gotHeight := headerInfo.Height //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
		if gotHeight != height {
			t.Errorf("Expected height %d got height %d", height, gotHeight)
		}

		gotWork := headerInfo.Work //nolint:staticcheck // TODO: look at the nil pointer issues mentioned by the linter here.
		if *gotWork != work {
			t.Errorf("Expected work %d got work %d", work.Uint64(), (*gotWork).Uint64())
		}
	})
}
