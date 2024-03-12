package types_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

func FuzzNewHeaderInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		// If either  of the hash strings is not of appropriate length
		// or not valid hex, generate a random hex randomly
		r := rand.New(rand.NewSource(seed))
		// Get the expected header bytes
		expectedHeaderBytes := datagen.GenRandomBTCHeaderInfo(r).Header
		expectedHeaderHashBytes := expectedHeaderBytes.Hash()
		height := datagen.GenRandomBTCHeight(r)
		work := sdkmath.NewUintFromBigInt(expectedHeaderBytes.Difficulty())

		headerInfo := types.NewBTCHeaderInfo(expectedHeaderBytes, expectedHeaderHashBytes, height, &work)
		// Check that all attributes are properly set
		if headerInfo == nil {
			t.Fatalf("returned object is nil")
		}
		if headerInfo.Header == nil {
			t.Errorf("Header inside header info is nil")
		}
		if headerInfo.Work == nil {
			t.Errorf("Work inside header info is nil")
		}

		gotHeaderBytes := headerInfo.Header.MustMarshal()
		if !bytes.Equal(expectedHeaderBytes.MustMarshal(), gotHeaderBytes) {
			t.Errorf("Expected header %v got %s", expectedHeaderBytes, gotHeaderBytes)
		}

		gotHashBytes := *headerInfo.Hash
		if !bytes.Equal(expectedHeaderHashBytes.MustMarshal(), gotHashBytes.MustMarshal()) {
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

func TestBTCHeaderInfoValidate(t *testing.T) {
	r := rand.New(rand.NewSource(120))
	valid := *datagen.GenRandomBTCHeaderInfo(r)
	valid2 := *datagen.GenRandomBTCHeaderInfo(r)
	zeroUint := sdkmath.ZeroUint()

	tcs := []struct {
		title       string
		info        types.BTCHeaderInfo
		expectedErr error
	}{
		{
			"valid",
			valid,
			nil,
		},
		{
			"valid2",
			valid2,
			nil,
		},
		{
			"invalid: header nil",
			types.BTCHeaderInfo{},
			errors.New("header is nil"),
		},
		{
			"invalid: hash nil",
			types.BTCHeaderInfo{
				Header: valid.Header,
			},
			errors.New("hash is nil"),
		},
		{
			"invalid: work nil",
			types.BTCHeaderInfo{
				Header: valid.Header,
				Hash:   valid.Hash,
			},
			errors.New("work is nil"),
		},
		{
			"invalid: height is zero",
			types.BTCHeaderInfo{
				Header: valid.Header,
				Hash:   valid.Hash,
				Work:   valid.Work,
				Height: 0,
			},
			errors.New("height is zero"),
		},
		{
			"invalid: height is zero",
			types.BTCHeaderInfo{
				Header: valid.Header,
				Hash:   valid.Hash,
				Work:   &zeroUint,
				Height: valid.Height,
			},
			errors.New("work is zero"),
		},
		{
			"invalid: bad block header",
			types.BTCHeaderInfo{
				Header: &bbn.BTCHeaderBytes{byte(1)},
				Hash:   valid.Hash,
				Work:   valid.Work,
				Height: valid.Height,
			},
			errors.New("unexpected EOF"),
		},
		{
			"invalid: bad block header",
			types.BTCHeaderInfo{
				Header: valid.Header,
				Hash:   valid2.Hash,
				Work:   valid.Work,
				Height: valid.Height,
			},
			fmt.Errorf("BTC header hash is not equal to generated hash from header %s != %s", valid2.Hash, valid.Hash),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.title, func(t *testing.T) {
			actErr := tc.info.Validate()
			if tc.expectedErr != nil {
				require.EqualError(t, actErr, tc.expectedErr.Error())
				return
			}
			require.NoError(t, actErr)
		})
	}
}
