package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochChainInfo gets the latest chain info of a given epoch for a given chain ID
func (k Keeper) GetEpochChainInfo(ctx sdk.Context, chainID string, epochNumber uint64) (*types.ChainInfo, error) {
	store := k.canonicalChainStore(ctx, chainID)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	if !store.Has(epochNumberBytes) {
		return nil, types.ErrEpochChainInfoNotFound
	}
	epochChainInfoBytes := store.Get(epochNumberBytes)
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(epochChainInfoBytes, &chainInfo)
	return &chainInfo, nil
}

// RecordEpochChainInfo records the chain info for a given epoch number of given chain ID
// where the latest chain info is retrieved from the chain info indexer
func (k Keeper) RecordEpochChainInfo(ctx sdk.Context, chainID string, epochNumber uint64) error {
	// get the latest known chain info
	chainInfo := k.GetChainInfo(ctx, chainID)
	// NOTE: we can record epoch chain info without ancestor since IBC connection can be established at any height
	store := k.epochChainInfoStore(ctx, chainID)
	store.Set(sdk.Uint64ToBigEndian(epochNumber), k.cdc.MustMarshal(chainInfo))
	return nil
}

// epochChainInfoStore stores each epoch's latest ChainInfo for a CZ
// prefix: EpochChainInfoKey || chainID
// key: epochNumber
// value: ChainInfo
func (k Keeper) epochChainInfoStore(ctx sdk.Context, chainID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	epochChainInfoStore := prefix.NewStore(store, types.EpochChainInfoKey)
	chainIDBytes := []byte(chainID)
	return prefix.NewStore(epochChainInfoStore, chainIDBytes)
}

// getFinalizedEpoch gets the last finalised epoch
// used upon querying the last BTC-finalised chain info for CZs
func (k Keeper) getFinalizedEpoch(ctx sdk.Context) (uint64, error) {
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
