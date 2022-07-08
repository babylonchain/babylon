package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"math/rand"
	"testing"
)

func FuzzNewHeaderInfo(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	btcdHeader := defaultHeader.ToBlockHeader()
	f.Add(
		btcdHeader.Version,
		btcdHeader.Bits,
		btcdHeader.Nonce,
		btcdHeader.Timestamp.Unix(),
		btcdHeader.PrevBlock.String(),
		btcdHeader.MerkleRoot.String(),
		int64(17))

	f.Fuzz(func(t *testing.T, version int32, bits uint32, nonce uint32,
		timeInt int64, prevBlockStr string, merkleRootStr string, seed int64) {

		// If either  of the hash strings is not of appropriate length
		// or not valid hex, generate a random hex randomly
		rand.Seed(seed)
		header := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, prevBlockStr, merkleRootStr)

		// Get the expected header bytes
		expectedHeaderBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

		headerInfo := types.NewHeaderInfo(header)
		if headerInfo == nil {
			t.Errorf("returned object is nil")
		}

		gotHeaderBytes := *headerInfo.Header
		if bytes.Compare(expectedHeaderBytes, gotHeaderBytes) != 0 {
			t.Errorf("Expected header %s got %s", expectedHeaderBytes, gotHeaderBytes)
		}

		gotHashBytes := *headerInfo.Hash
		blockHash := header.BlockHash()
		expectedHashBytes := bbl.NewBTCHeaderHashBytesFromChainhash(&blockHash)
		if bytes.Compare(expectedHashBytes, gotHashBytes) != 0 {
			t.Errorf("Expected header hash %s got %s", expectedHashBytes, gotHashBytes)
		}
	})
}
