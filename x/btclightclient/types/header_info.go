package types

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/wire"
)

func NewHeaderInfo(header *wire.BlockHeader) (*HeaderInfo, error) {
	headerHashCh := header.BlockHash()

	currentHeaderBytes, err := bbl.NewBTCHeaderBytesFromBlockHeader(header)
	if err != nil {
		return nil, err
	}
	headerHash, err := bbl.NewBTCHeaderHashBytesFromChainhash(&headerHashCh)
	if err != nil {
		return nil, err
	}
	return &HeaderInfo{
		Header: &currentHeaderBytes,
		Hash:   &headerHash,
	}, nil
}
