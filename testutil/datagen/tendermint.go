package datagen

import (
	"time"

	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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
