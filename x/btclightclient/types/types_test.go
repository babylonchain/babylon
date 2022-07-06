package types_test

import (
	"encoding/hex"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math/rand"
	"time"
)

func genRandomByteArray(length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	rand.Read(newHeaderBytes)
	return newHeaderBytes
}

func genRandomHexStr(length uint64) string {
	randBytes := genRandomByteArray(length)
	return hex.EncodeToString(randBytes)
}

func genRandomBtcdHeader(version int32, bits uint32, nonce uint32,
	timeInt int64, prevBlockStr string, merkleRootStr string) *wire.BlockHeader {
	if !validHex(prevBlockStr, bbl.BTCHeaderHashLen) {
		prevBlockStr = genRandomHexStr(bbl.BTCHeaderHashLen)
	}
	if !validHex(merkleRootStr, bbl.BTCHeaderHashLen) {
		merkleRootStr = genRandomHexStr(bbl.BTCHeaderHashLen)
	}

	// Get the chainhash versions
	prevBlock, _ := chainhash.NewHashFromStr(prevBlockStr)
	merkleRoot, _ := chainhash.NewHashFromStr(merkleRootStr)
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

	return &header
}

// validHex accepts a hex string and the length representation as a byte array
func validHex(hexStr string, length int) bool {
	if len(hexStr) != length*2 {
		return false
	}
	if _, err := hex.DecodeString(hexStr); err != nil {
		return false
	}
	return true
}
