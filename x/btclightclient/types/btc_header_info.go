package types

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/wire"
)

func NewBTCHeaderInfo(header *wire.BlockHeader, height uint64) *BTCHeaderInfo {
	headerHashCh := header.BlockHash()

	currentHeaderBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

	headerHash := bbl.NewBTCHeaderHashBytesFromChainhash(&headerHashCh)
	return &BTCHeaderInfo{
		Header: &currentHeaderBytes,
		Hash:   &headerHash,
		Height: height,
	}
}
