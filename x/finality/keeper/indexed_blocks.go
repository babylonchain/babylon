package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IndexBlock indexes the current block, saves the corresponding indexed block
// to KVStore
func (k Keeper) IndexBlock(ctx context.Context) {
	headerInfo := sdk.UnwrapSDKContext(ctx).HeaderInfo()
	ib := &types.IndexedBlock{
		Height:    uint64(headerInfo.Height),
		AppHash:   headerInfo.AppHash,
		Finalized: false,
	}
	k.SetBlock(ctx, ib)

	// record the block height
	types.RecordLastHeight(int(headerInfo.Height))
}

func (k Keeper) SetBlock(ctx context.Context, block *types.IndexedBlock) {
	store := k.blockStore(ctx)
	blockBytes := k.cdc.MustMarshal(block)
	store.Set(sdk.Uint64ToBigEndian(block.Height), blockBytes)
}

func (k Keeper) HasBlock(ctx context.Context, height uint64) bool {
	store := k.blockStore(ctx)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetBlock(ctx context.Context, height uint64) (*types.IndexedBlock, error) {
	store := k.blockStore(ctx)
	blockBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(blockBytes) == 0 {
		return nil, types.ErrBlockNotFound
	}
	var block types.IndexedBlock
	k.cdc.MustUnmarshal(blockBytes, &block)
	return &block, nil
}

// blockStore returns the KVStore of the blocks
// prefix: BlockKey
// key: block height
// value: IndexedBlock
func (k Keeper) blockStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.BlockKey)
}
