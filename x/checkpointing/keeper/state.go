package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BlsSigsState struct {
	cdc         codec.BinaryCodec
	blsSigs     sdk.KVStore
	hashToEpoch sdk.KVStore
}

func (k Keeper) BlsSigsState(ctx sdk.Context) BlsSigsState {
	// Build the BlsSigsState storage
	store := ctx.KVStore(k.storeKey)
	return BlsSigsState{
		cdc:     k.cdc,
		blsSigs: prefix.NewStore(store, types.BlsSigsPrefix),
	}
}

// InsertBlsSig inserts a bls sig into storage
func (bs BlsSigsState) InsertBlsSig(sig *types.BlsSig) {
	epoch := sig.GetEpochNum()
	sigHash := sig.Hash()
	blsSigsKey := types.BlsSigsObjectKey(epoch, sigHash)
	epochKey := types.BlsSigsEpochKey(sigHash)

	bs.blsSigs.Set(blsSigsKey, bs.cdc.MustMarshal(sig))
	bs.hashToEpoch.Set(epochKey, sdk.Uint64ToBigEndian(epoch))
}

// GetBlsSig retrieves a bls sig by its epoch and hash
func (bs BlsSigsState) GetBlsSig(epoch uint64, hash types.BlsSigHash) (*types.BlsSig, error) {
	blsSigsKey := types.BlsSigsObjectKey(epoch, hash)
	bz := bs.blsSigs.Get(blsSigsKey)
	if bz == nil {
		return nil, types.ErrBlsSigDoesNotExist.Wrap("no header with provided height and hash")
	}

	blsSig := new(types.BlsSig)
	bs.cdc.MustUnmarshal(bz, blsSig)
	return blsSig, nil
}

// GetBlsSigByHash retrieves a bls sig by its hash
func (bs BlsSigsState) GetBlsSigByHash(hash types.BlsSigHash) (*types.BlsSig, error) {
	epoch, err := bs.GetBlsSigEpoch(hash)
	if err != nil {
		return nil, err
	}
	return bs.GetBlsSig(epoch, hash)
}

// GetBlsSigsByEpoch retrieves bls sigs by their epoch
func (bs BlsSigsState) GetBlsSigsByEpoch(epoch uint64) ([]*types.BlsSig, error) {
	store := prefix.NewStore(bs.blsSigs, sdk.Uint64ToBigEndian(epoch))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	blsSigs := make([]*types.BlsSig, 0)
	for ; iter.Valid(); iter.Next() {
		rawBytes := iter.Value()
		blsSig := new(types.BlsSig)
		bs.cdc.MustUnmarshal(rawBytes, blsSig)
		blsSigs = append(blsSigs, blsSig)
	}
	if len(blsSigs) == 0 {
		return nil, types.ErrBlsSigsEpochDoesNotExist.Wrap("no bls sigs with provided epoch")
	}
	return blsSigs, nil
}

func (bs BlsSigsState) GetBlsSigEpoch(hash types.BlsSigHash) (uint64, error) {
	hashKey := types.BlsSigsEpochKey(hash)
	bz := bs.hashToEpoch.Get(hashKey)
	if bz == nil {
		return 0, types.ErrBlsSigDoesNotExist.Wrap("no bls sig with provided hash")
	}
	return sdk.BigEndianToUint64(bz), nil
}

// Exists Check whether a hash is maintained in storage
func (bs BlsSigsState) Exists(hash types.BlsSigHash) bool {
	_, err := bs.GetBlsSigEpoch(hash)
	return err == nil
}

type CheckpointsState struct {
	cdc         codec.BinaryCodec
	checkpoints sdk.KVStore
}

func (k Keeper) CheckpointsState(ctx sdk.Context) CheckpointsState {
	// Build the CheckpointsState storage
	store := ctx.KVStore(k.storeKey)
	return CheckpointsState{
		cdc:         k.cdc,
		checkpoints: prefix.NewStore(store, types.CheckpointsPrefix),
	}
}
