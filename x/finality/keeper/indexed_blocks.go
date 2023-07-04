package keeper

import (
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IndexBlock indexes the current block, saves the corresponding indexed block
// to KVStore
func (k Keeper) IndexBlock(ctx sdk.Context) {
	header := ctx.BlockHeader()
	ib := &types.IndexedBlock{
		Height:         uint64(header.Height),
		LastCommitHash: header.LastCommitHash,
		Finalized:      false,
	}
	k.setBlock(ctx, ib)
}

func (k Keeper) setBlock(ctx sdk.Context, block *types.IndexedBlock) {
	store := k.blockStore(ctx)
	blockBytes := k.cdc.MustMarshal(block)
	store.Set(sdk.Uint64ToBigEndian(block.Height), blockBytes)
}

func (k Keeper) HasBlock(ctx sdk.Context, height uint64) bool {
	store := k.blockStore(ctx)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetBlock(ctx sdk.Context, height uint64) (*types.IndexedBlock, error) {
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
func (k Keeper) blockStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.BlockKey)
}
