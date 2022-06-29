package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultEpochNumber = 0
)

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		return sdk.NewUint(uint64(DefaultEpochNumber)), nil
	}
	var epochNumber sdk.Uint
	err := epochNumber.Unmarshal(bz)

	return epochNumber, err
}

// setEpochNumber sets epoch number
func (k Keeper) setEpochNumber(ctx sdk.Context, epochNumber sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)

	return nil
}

// IncEpochNumber adds epoch number by 1
func (k Keeper) IncEpochNumber(ctx sdk.Context) error {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return err
	}
	incrementedEpochNumber := epochNumber.AddUint64(1)
	return k.setEpochNumber(ctx, incrementedEpochNumber)
}

// GetEpochBoundary gets the epoch boundary, i.e., the height of the block that ends this epoch
func (k Keeper) GetEpochBoundary(ctx sdk.Context) (sdk.Uint, error) {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return sdk.NewUint(0), err
	}
	epochInterval := sdk.NewUint(k.GetParams(ctx).EpochInterval)
	// example: in epoch 0, epoch interval is 5 blocks, boundary will be (0+1)*5=5
	return epochNumber.AddUint64(1).Mul(epochInterval), nil
}
