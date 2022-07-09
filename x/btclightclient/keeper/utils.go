package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func blockHeaderFromStoredBytes(bz []byte) *wire.BlockHeader {
	// Convert the bytes value into a BTCHeaderBytes object
	headerBytes, err := bbl.NewBTCHeaderBytesFromBytes(bz)
	if err != nil {
		panic("Stored bytes cannot be converted to BTCHeaderBytes object")
	}
	// Convert the BTCHeaderBytes object into a *wire.BlockHeader object
	return headerBytes.ToBlockHeader()
}

func isParent(child *wire.BlockHeader, parent *wire.BlockHeader) bool {
	return sameHash(child.PrevBlock, parent.BlockHash())
}

func sameBlock(header1 *wire.BlockHeader, header2 *wire.BlockHeader) bool {
	return sameHash(header1.BlockHash(), header2.BlockHash())
}

func sameHash(hash1 chainhash.Hash, hash2 chainhash.Hash) bool {
	return hash1.String() == hash2.String()
}
