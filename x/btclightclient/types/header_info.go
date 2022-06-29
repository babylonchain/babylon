package types

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/wire"
)

func NewHeaderInfo(header *wire.BlockHeader) *HeaderInfo {
	currentHeaderBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)
	headerHash := bbl.NewBTCHeaderHashBytesFromChainhash(header.BlockHash())
	return &HeaderInfo{
		Header: &currentHeaderBytes,
		Hash:   &headerHash,
	}
}
