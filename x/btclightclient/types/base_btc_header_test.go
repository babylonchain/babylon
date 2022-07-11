package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"math/rand"
	"testing"
)

func FuzzBaseBTCHeader(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	defaultBtcdHeader := defaultHeader.ToBlockHeader()

	f.Add(
		defaultBtcdHeader.Version,
		defaultBtcdHeader.Bits,
		defaultBtcdHeader.Nonce,
		defaultBtcdHeader.Timestamp.Unix(),
		defaultBtcdHeader.PrevBlock.String(),
		defaultBtcdHeader.MerkleRoot.String(),
		uint64(42),
		int64(17))

	f.Fuzz(func(t *testing.T, version int32, bits uint32, nonce uint32,
		timeInt int64, prevBlockStr string, merkleRootStr string, height uint64, seed int64) {

		rand.Seed(seed)
		// Get the btcd header based on the provided data
		btcdHeader := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, prevBlockStr, merkleRootStr)
		// Convert it into bytes
		headerBytesObj := bbl.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
		headerBytes, _ := headerBytesObj.Marshal()

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
