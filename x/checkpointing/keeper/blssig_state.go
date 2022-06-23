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
		cdc:         k.cdc,
		blsSigs:     prefix.NewStore(store, types.BlsSigsPrefix),
		hashToEpoch: prefix.NewStore(store, types.BlsSigsHashToEpochPrefix),
	}
}

// CreateBlsSig inserts the bls sig into the hash->epoch and (epoch, hash)->bls sig storage
func (bs BlsSigsState) CreateBlsSig(sig *types.BlsSig) {
	epoch := sig.GetEpochNum()
	sigHash := sig.Hash()
	blsSigsKey := types.BlsSigsObjectKey(epoch, sigHash)
	epochKey := types.BlsSigsEpochKey(sigHash)

	// save concrete bls sig object
	bs.blsSigs.Set(blsSigsKey, bs.cdc.MustMarshal(sig))
	// map bls sig to epoch
	bs.hashToEpoch.Set(epochKey, sdk.Uint64ToBigEndian(epoch))
}

// GetBlsSig retrieves a bls sig by its epoch and hash
func (bs BlsSigsState) GetBlsSig(epoch uint64, hash types.BlsSigHash) (*types.BlsSig, error) {
	blsSigsKey := types.BlsSigsObjectKey(epoch, hash)
	rawBytes := bs.blsSigs.Get(blsSigsKey)
	if rawBytes == nil {
		return nil, types.ErrBlsSigDoesNotExist.Wrap("no header with provided epoch and hash")
	}

	blsSig := new(types.BlsSig)
	bs.cdc.MustUnmarshal(rawBytes, blsSig)
	return blsSig, nil
}

// GetBlsSigEpoch retrieves the epoch of a bls sig
func (bs BlsSigsState) GetBlsSigEpoch(hash types.BlsSigHash) (uint64, error) {
	hashKey := types.BlsSigsEpochKey(hash)
	bz := bs.hashToEpoch.Get(hashKey)
	if bz == nil {
		return 0, types.ErrBlsSigDoesNotExist.Wrap("no bls sig with provided hash")
	}
	return sdk.BigEndianToUint64(bz), nil
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
func (bs BlsSigsState) GetBlsSigsByEpoch(epoch uint64, f func(sig *types.BlsSig) bool) error {
	store := prefix.NewStore(bs.blsSigs, sdk.Uint64ToBigEndian(epoch))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		rawBytes := iter.Value()
		blsSig := new(types.BlsSig)
		bs.cdc.MustUnmarshal(rawBytes, blsSig)
		stop := f(blsSig)
		if stop {
			break
		}
	}
	return nil
}

// Exists Check whether a hash is maintained in storage
func (bs BlsSigsState) Exists(hash types.BlsSigHash) bool {
	return bs.hashToEpoch.Has(types.BlsSigHashToBytes(hash))
}
