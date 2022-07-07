package types

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/wire"
)

func NewHeaderInfo(header *wire.BlockHeader) *HeaderInfo {
	headerHashCh := header.BlockHash()

	currentHeaderBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

	headerHash := bbl.NewBTCHeaderHashBytesFromChainhash(&headerHashCh)
	return &HeaderInfo{
		Header: &currentHeaderBytes,
		Hash:   &headerHash,
	}
}
