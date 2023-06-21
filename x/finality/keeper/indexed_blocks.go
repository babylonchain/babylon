package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddBlock(ctx sdk.Context, block *types.IndexedBlock) error {
	if k.HasBlock(ctx, block.Height) {
		block2, err := k.GetBlock(ctx, block.Height)
		if err != nil {
			panic(err)
		}
		if !block.Equal(block2) {
			panic(fmt.Errorf("conflicting blocks detected at height %d", block.Height))
		}
		return types.ErrDuplicatedBlock
	}
	k.setBlock(ctx, block)
	return nil
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
