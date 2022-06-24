package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetBaseBTCHeader(ctx sdk.Context) types.BaseBTCHeader {
	baseBtcdHeader, err := k.HeadersState(ctx).GetBaseBTCHeader()

	if err != nil {
		return types.BaseBTCHeader{}
	}

	if baseBtcdHeader == nil {
		return types.BaseBTCHeader{}
	}

	baseHash := baseBtcdHeader.BlockHash()
	height, err := k.HeadersState(ctx).GetHeaderHeight(&baseHash)

	if err != nil {
		return types.BaseBTCHeader{}
	}

	var headerBytes bbl.BTCHeaderBytes
	headerBytes.UnmarshalBlockHeader(baseBtcdHeader)
	return types.BaseBTCHeader{Header: &headerBytes, Height: height}
}

// SetBaseBTCHeader checks whether a base BTC header exists and
// 					if not inserts it into storage
func (k Keeper) SetBaseBTCHeader(ctx sdk.Context, baseBTCHeader types.BaseBTCHeader) {
	existingHeader, _ := k.HeadersState(ctx).GetBaseBTCHeader()
	if existingHeader != nil {
		panic("A base BTC Header has already been set")
	}

	btcdHeader, err := baseBTCHeader.Header.MarshalBlockHeader()
	if err != nil {
		panic("Base BTC Header bytes do not correspond to btcd header")
	}
	k.HeadersState(ctx).CreateHeader(btcdHeader, baseBTCHeader.Height)
}
