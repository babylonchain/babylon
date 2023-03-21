package datagen

import (
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func GenRandomTMHeader(chainID string, height uint64) *tmproto.Header {
	return &tmproto.Header{
		ChainID:        chainID,
		Height:         int64(height),
		Time:           time.Now(),
		LastCommitHash: GenRandomByteArray(32),
	}
}

func GenRandomIBCTMHeader(chainID string, height uint64) *ibctmtypes.Header {
	return &ibctmtypes.Header{
		SignedHeader: &tmproto.SignedHeader{
			Header: &tmproto.Header{
				ChainID:        chainID,
				Height:         int64(height),
				LastCommitHash: GenRandomByteArray(32),
			},
		},
	}
}

func HeaderToHeaderInfo(header *ibctmtypes.Header) *ibcclientkeeper.HeaderInfo {
	return &ibcclientkeeper.HeaderInfo{
		Hash:     header.Header.LastCommitHash,
		ChaindId: header.Header.ChainID,
		Height:   uint64(header.Header.Height),
	}
}
