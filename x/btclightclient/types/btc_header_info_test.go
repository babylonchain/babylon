package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzNewHeaderInfo(f *testing.F) {
	defaultHeader := bbl.GetBaseHeaderBytes()
	btcdHeader := defaultHeader.ToBlockHeader()
	f.Add(
		btcdHeader.Version,
		btcdHeader.Bits,
		btcdHeader.Nonce,
		btcdHeader.Timestamp.Unix(),
		btcdHeader.PrevBlock.String(),
		btcdHeader.MerkleRoot.String(),
		uint64(42),
		uint64(24),
		int64(17))

	f.Fuzz(func(t *testing.T, version int32, bits uint32, nonce uint32,
		timeInt int64, prevBlockStr string, merkleRootStr string, height uint64, work uint64, seed int64) {

		// If either  of the hash strings is not of appropriate length
		// or not valid hex, generate a random hex randomly
		rand.Seed(seed)
		header := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, prevBlockStr, merkleRootStr)

		workSdk := sdk.NewUint(work)

		// Get the expected header bytes
		expectedHeaderBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)
		expectedHeaderHashBytes := expectedHeaderBytes.Hash()

		headerInfo := types.NewBTCHeaderInfo(&expectedHeaderBytes, expectedHeaderHashBytes, height, &workSdk)
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

		gotHeight := headerInfo.Height
		if gotHeight != height {
			t.Errorf("Expected height %d got height %d", height, gotHeight)
		}

		gotWork := headerInfo.Work
		if *gotWork != workSdk {
			t.Errorf("Expected work %d got work %d", workSdk.Uint64(), (*gotWork).Uint64())
		}
	})
}
