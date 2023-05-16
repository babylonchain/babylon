package keeper

import (
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetFinalizingBTCTip gets the BTC tip when the last epoch is finalised
func (k Keeper) GetFinalizingBTCTip(ctx sdk.Context) *btclctypes.BTCHeaderInfo {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.FinalizingBTCTipKey) {
		return nil
	}
	btcTipBytes := store.Get(types.FinalizingBTCTipKey)
	var btcTip btclctypes.BTCHeaderInfo
	k.cdc.MustUnmarshal(btcTipBytes, &btcTip)
	return &btcTip
}

// setFinalizingBTCTip sets the last finalised BTC tip
// called upon each AfterRawCheckpointFinalized hook invocation
func (k Keeper) setFinalizingBTCTip(ctx sdk.Context, btcTip *btclctypes.BTCHeaderInfo) {
	store := ctx.KVStore(k.storeKey)
	btcTipBytes := k.cdc.MustMarshal(btcTip)
	store.Set(types.FinalizingBTCTipKey, btcTipBytes)
}

// GetFinalizedEpoch gets the last finalised epoch
// used upon querying the last BTC-finalised chain info for CZs
func (k Keeper) GetFinalizedEpoch(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.FinalizedEpochKey) {
		return 0, types.ErrFinalizedEpochNotFound
	}
	epochNumberBytes := store.Get(types.FinalizedEpochKey)
	return sdk.BigEndianToUint64(epochNumberBytes), nil
}

// setFinalizedEpoch sets the last finalised epoch
// called upon each AfterRawCheckpointFinalized hook invocation
func (k Keeper) setFinalizedEpoch(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Set(types.FinalizedEpochKey, epochNumberBytes)
}

func (k Keeper) GetEpoch(ctx sdk.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}
