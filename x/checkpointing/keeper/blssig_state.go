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
	bs.blsSigs.Set(blsSigsKey, types.BlsSigToBytes(bs.cdc, sig))
	// map bls sig to epoch
	bs.hashToEpoch.Set(epochKey, sdk.Uint64ToBigEndian(epoch))
}

// GetBlsSigsByEpoch retrieves bls sigs by their epoch
func (bs BlsSigsState) GetBlsSigsByEpoch(epoch uint64, f func(sig *types.BlsSig) bool) error {
	store := prefix.NewStore(bs.blsSigs, sdk.Uint64ToBigEndian(epoch))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		rawBytes := iter.Value()
		blsSig, err := types.BytesToBlsSig(bs.cdc, rawBytes)
		if err != nil {
			return err
		}
		stop := f(blsSig)
		if stop {
			return nil
		}
	}
	return nil
}

// Exists Check whether a bls sig is maintained in storage
func (bs BlsSigsState) Exists(hash types.BlsSigHash) bool {
	store := prefix.NewStore(bs.hashToEpoch, types.BlsSigsHashToEpochPrefix)
	return store.Has(hash.Bytes())
}
