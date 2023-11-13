package datagen

import (
	"math/rand"
	"time"

	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func GenRandomTMHeader(r *rand.Rand, chainID string, height uint64) *tmproto.Header {
	return &tmproto.Header{
		ChainID:        chainID,
		Height:         int64(height),
		Time:           time.Now(),
		LastCommitHash: GenRandomByteArray(r, 32),
	}
}

func GenRandomIBCTMHeader(r *rand.Rand, chainID string, height uint64) *ibctmtypes.Header {
	return &ibctmtypes.Header{
		SignedHeader: &tmproto.SignedHeader{
			Header: &tmproto.Header{
				ChainID:        chainID,
				Height:         int64(height),
				LastCommitHash: GenRandomByteArray(r, 32),
			},
		},
	}
}

func HeaderToHeaderInfo(header *ibctmtypes.Header) *zctypes.HeaderInfo {
	return &zctypes.HeaderInfo{
		Hash:    header.Header.LastCommitHash,
		ChainId: header.Header.ChainID,
		Height:  uint64(header.Header.Height),
	}
}
