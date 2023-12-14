package datagen

import (
	"math/rand"
	"time"

	"cosmossdk.io/core/header"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

func GenRandomTMHeader(r *rand.Rand, chainID string, height uint64) *tmproto.Header {
	return &tmproto.Header{
		ChainID: chainID,
		Height:  int64(height),
		Time:    time.Now(),
		AppHash: GenRandomByteArray(r, 32),
	}
}

func GenRandomIBCTMHeader(r *rand.Rand, chainID string, height uint64) *ibctmtypes.Header {
	return &ibctmtypes.Header{
		SignedHeader: &tmproto.SignedHeader{
			Header: &tmproto.Header{
				ChainID: chainID,
				Height:  int64(height),
				AppHash: GenRandomByteArray(r, 32),
			},
		},
	}
}

func GenRandomTMHeaderInfo(r *rand.Rand, chainID string, height uint64) *header.Info {
	tmHeader := GenRandomIBCTMHeader(r, chainID, height)
	return &header.Info{
		Height:  tmHeader.Header.Height,
		Hash:    tmHeader.Header.DataHash,
		Time:    tmHeader.Header.Time,
		ChainID: tmHeader.Header.ChainID,
		AppHash: tmHeader.Header.AppHash,
	}
}

func HeaderToHeaderInfo(header *ibctmtypes.Header) *zctypes.HeaderInfo {
	return &zctypes.HeaderInfo{
		AppHash: header.Header.AppHash,
		ChainId: header.Header.ChainID,
		Time:    header.Header.Time,
		Height:  uint64(header.Header.Height),
	}
}

func WithCtxHeight(ctx sdk.Context, height uint64) sdk.Context {
	headerInfo := ctx.HeaderInfo()
	headerInfo.Height = int64(height)
	ctx = ctx.WithHeaderInfo(headerInfo)
	return ctx
}
