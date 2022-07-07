package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetBaseBTCHeader(ctx sdk.Context) types.BaseBTCHeader {
	baseBtcdHeader := k.HeadersState(ctx).GetBaseBTCHeader()

	if baseBtcdHeader == nil {
		return types.BaseBTCHeader{}
	}

	baseHash := baseBtcdHeader.BlockHash()
	height, err := k.HeadersState(ctx).GetHeaderHeight(&baseHash)

	if err != nil {
		return types.BaseBTCHeader{}
	}

	headerBytes := bbl.NewBTCHeaderBytesFromBlockHeader(baseBtcdHeader)
	return types.BaseBTCHeader{Header: &headerBytes, Height: height}
}

// SetBaseBTCHeader checks whether a base BTC header exists and
// 					if not inserts it into storage
func (k Keeper) SetBaseBTCHeader(ctx sdk.Context, baseBTCHeader types.BaseBTCHeader) {
	existingHeader := k.HeadersState(ctx).GetBaseBTCHeader()
	if existingHeader != nil {
		panic("A base BTC Header has already been set")
	}

	btcdHeader := baseBTCHeader.Header.ToBlockHeader()

	// The cumulative work for the Base BTC header is only the work
	// for that particular header. This means that it is very important
	// that no forks will happen that discard the base header because we
	// will not be able to detect those. Cumulative work will build based
	// on the sum of the work of the chain starting from the base header.
	blockWork := types.CalcWork(btcdHeader)

	k.HeadersState(ctx).CreateHeader(btcdHeader, baseBTCHeader.Height, blockWork)
}
