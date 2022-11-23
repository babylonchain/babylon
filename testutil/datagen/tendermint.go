package datagen

import (
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

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
