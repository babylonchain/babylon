package datagen

import (
	"encoding/hex"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math/rand"
	"time"
)

func GenRandomByteArray(length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	rand.Read(newHeaderBytes)
	return newHeaderBytes
}

func GenRandomHexStr(length uint64) string {
	randBytes := GenRandomByteArray(length)
	return hex.EncodeToString(randBytes)
}

func GenRandomBtcdHeader(version int32, bits uint32, nonce uint32,
	timeInt int64, prevBlockStr string, merkleRootStr string) *wire.BlockHeader {
	if !ValidHex(prevBlockStr, bbl.BTCHeaderHashLen) {
		prevBlockStr = GenRandomHexStr(bbl.BTCHeaderHashLen)
	}
	if !ValidHex(merkleRootStr, bbl.BTCHeaderHashLen) {
		merkleRootStr = GenRandomHexStr(bbl.BTCHeaderHashLen)
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

// ValidHex accepts a hex string and the length representation as a byte array
func ValidHex(hexStr string, length int) bool {
	if len(hexStr) != length*2 {
		return false
	}
	if _, err := hex.DecodeString(hexStr); err != nil {
		return false
	}
	return true
}
