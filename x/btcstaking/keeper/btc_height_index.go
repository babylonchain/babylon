package keeper

import (
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IndexBTCHeight indexes the current BTC height, and saves it to KVStore
func (k Keeper) IndexBTCHeight(ctx sdk.Context) {
	babylonHeight := uint64(ctx.BlockHeight())
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		return
	}
	btcHeight := btcTip.Height
	store := k.btcHeightStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(babylonHeight), sdk.Uint64ToBigEndian(btcHeight))
}

func (k Keeper) GetBTCHeightAtBabylonHeight(ctx sdk.Context, babylonHeight uint64) (uint64, error) {
	store := k.btcHeightStore(ctx)
	btcHeightBytes := store.Get(sdk.Uint64ToBigEndian(babylonHeight))
	if len(btcHeightBytes) == 0 {
		return 0, types.ErrBTCHeightNotFound
	}
	return sdk.BigEndianToUint64(btcHeightBytes), nil
}

func (k Keeper) GetCurrentBTCHeight(ctx sdk.Context) (uint64, error) {
	babylonHeight := uint64(ctx.BlockHeight())
	return k.GetBTCHeightAtBabylonHeight(ctx, babylonHeight)
}

// btcHeightStore returns the KVStore of the BTC heights
// prefix: BTCHeightKey
// key: Babylon block height
// value: BTC block height
func (k Keeper) btcHeightStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.BTCHeightKey)
}
