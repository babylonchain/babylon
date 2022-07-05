package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"math/rand"
	"testing"
)

func FuzzBaseBTCHeader(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	defaultHeaderBytes, _ := defaultHeader.Marshal()
	f.Add(defaultHeaderBytes, uint64(42), int64(17))

	f.Fuzz(func(t *testing.T, headerBytes []byte, height uint64, seed int64) {
		// HeaderBytes should have a length of bbl.HeaderLen
		// If not, generate a random array using of that length using the seed
		if len(headerBytes) != bbl.BTCHeaderLen {
			rand.Seed(seed)
			headerBytes = genRandomByteArray(bbl.BTCHeaderLen)
		}
		headerBytesObj, err := bbl.NewBTCHeaderBytesFromBytes(headerBytes)
		if err != nil {
			// Invalid header bytes
			t.Skip()
		}

		baseBTCHeader := types.NewBaseBTCHeader(headerBytesObj, height)

		// Validate the various attributes
		gotHeaderBytes, err := baseBTCHeader.Header.Marshal()
		if err != nil {
			t.Errorf("Header returned cannot be marshalled")
		}
		gotHeight := baseBTCHeader.Height

		if bytes.Compare(headerBytes, gotHeaderBytes) != 0 {
			t.Errorf("Header attribute is different")
		}
		if height != gotHeight {
			t.Errorf("Height attribute is different")
		}

		// Perform the header validation
		err = baseBTCHeader.Validate()
		if err != nil {
			t.Errorf("Got error %s when validating", err)
		}
	})
}
