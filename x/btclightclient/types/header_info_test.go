package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math/rand"
	"testing"
	"time"
)

func FuzzNewHeaderInfo(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	btcdHeader, _ := defaultHeader.ToBlockHeader()
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
		if !validHex(prevBlockStr, bbl.BTCHeaderHashLen) {
			prevBlockStr = genRandomHexStr(bbl.BTCHeaderHashLen)
		}
		if !validHex(merkleRootStr, bbl.BTCHeaderHashLen) {
			merkleRootStr = genRandomHexStr(bbl.BTCHeaderHashLen)
		}

		// Get the chainhash versions
		prevBlock, err := chainhash.NewHashFromStr(prevBlockStr)
		if err != nil {
			t.Skip()
		}
		merkleRoot, err := chainhash.NewHashFromStr(merkleRootStr)
		if err != nil {
			t.Skip()
		}
		time := time.Unix(timeInt, 0)

		// Construct a header
		header := wire.BlockHeader{
			Version:    version,
			Bits:       bits,
			Nonce:      nonce,
			PrevBlock:  *prevBlock,
			MerkleRoot: *merkleRoot,
			Timestamp:  time,
		}

		headerInfo := types.NewHeaderInfo(&header)

		if headerInfo == nil {
			t.Errorf("returned object is nil")
		}

		gotHeader := *headerInfo.Header
		expectedHeader := bbl.NewBTCHeaderBytesFromBlockHeader(&header)
		if bytes.Compare(expectedHeader, gotHeader) != 0 {
			t.Errorf("Expected header %s got %s", expectedHeader, gotHeader)
		}

		gotHash := *headerInfo.Hash
		blockHash := header.BlockHash()
		expectedHash := bbl.NewBTCHeaderHashBytesFromChainhash(&blockHash)
		if bytes.Compare(expectedHash, gotHash) != 0 {
			t.Errorf("Expected header hash %s got %s", expectedHash, gotHash)
		}
	})
}
