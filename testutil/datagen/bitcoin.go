package datagen

import (
	bbl "github.com/babylonchain/babylon/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math/rand"
	"time"
)

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

func GenRandomHeaderInfo() *btclightclienttypes.BTCHeaderInfo {
	header, _ := bbl.NewBTCHeaderBytesFromBytes(GenRandomByteArray(bbl.BTCHeaderLen))
	headerHash := header.Hash()
	height := rand.Uint64()
	work := btclightclienttypes.CalcWork(&header)

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &header,
		Hash:   headerHash,
		Height: height,
		Work:   &work,
	}
}

func GenRandomHeaderInfoWithHeight(height uint64) *btclightclienttypes.BTCHeaderInfo {
	headerInfo := GenRandomHeaderInfo()
	headerInfo.Height = height
	return headerInfo
}

func MutateHash(hash *bbl.BTCHeaderHashBytes) *bbl.BTCHeaderHashBytes {
	mutatedBytes := make([]byte, bbl.BTCHeaderHashLen)
	copy(mutatedBytes, hash.MustMarshal())
	mutatedBytes[0] -= 1
	mutated, _ := bbl.NewBTCHeaderHashBytesFromBytes(mutatedBytes)
	return &mutated
}
