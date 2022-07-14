package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetBaseBTCHeader(ctx sdk.Context) *types.BTCHeaderInfo {
	return k.headersState(ctx).GetBaseBTCHeader()
}

// SetBaseBTCHeader checks whether a base BTC header exists and
// 					if not inserts it into storage
func (k Keeper) SetBaseBTCHeader(ctx sdk.Context, baseBTCHeader types.BTCHeaderInfo) {
	existingHeader := k.headersState(ctx).GetBaseBTCHeader()
	if existingHeader != nil {
		panic("A base BTC Header has already been set")
	}
	k.headersState(ctx).CreateHeader(&baseBTCHeader)
}
