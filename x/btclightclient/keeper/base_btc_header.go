package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetBaseBTCHeader(ctx sdk.Context) types.BTCHeaderInfo {
	baseBtcHeader := k.HeadersState(ctx).GetBaseBTCHeader()
	return *baseBtcHeader
}

// SetBaseBTCHeader checks whether a base BTC header exists and
// 					if not inserts it into storage
func (k Keeper) SetBaseBTCHeader(ctx sdk.Context, baseBTCHeader types.BTCHeaderInfo) {
	existingHeader := k.HeadersState(ctx).GetBaseBTCHeader()
	if existingHeader != nil {
		panic("A base BTC Header has already been set")
	}
	k.HeadersState(ctx).CreateHeader(&baseBTCHeader)
}
