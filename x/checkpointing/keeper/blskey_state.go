package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BlsKeysState struct {
	cdc     codec.BinaryCodec
	blsKeys sdk.KVStore
}

func (k Keeper) BlsKeysState(ctx sdk.Context) BlsKeysState {
	// Build the BlsKeysState storage
	store := ctx.KVStore(k.storeKey)
	return BlsKeysState{
		cdc:     k.cdc,
		blsKeys: prefix.NewStore(store, types.BlsKeysPrefix),
	}
}

// CreateBlsKey inserts the BLS key into the address->bls_key storage
func (bk BlsKeysState) CreateBlsKey(key *types.BlsPubKey) error {
	if bk.Exists(key.Address) {
		return types.ErrBlsKeyAlreadyExist.Wrapf("existed public key: %x", key.Key)
	}
	blsKeysKey := types.BlsKeysObjectKey(key.Address)
	// save concrete BLS public key object
	bk.blsKeys.Set(blsKeysKey, types.BlsPubKeyToBytes(bk.cdc, key))
	return nil
}

// GetBlsPubKey retrieves BLS public key by validator's address
func (bk BlsKeysState) GetBlsPubKey(addr string) (*types.BlsPubKey, error) {
	pubKeyKey := types.BlsKeysObjectKey(addr)
	rawBytes := bk.blsKeys.Get(pubKeyKey)
	if rawBytes == nil {
		return nil, types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}

	return types.BytesToBlsPubKey(bk.cdc, rawBytes)
}

// RemoveBlsPubKey removes a BLS public key
func (bk BlsKeysState) RemoveBlsPubKey(addr string) error {
	if !bk.Exists(addr) {
		return types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}
	// delete BLS public key from storage
	bk.blsKeys.Delete(types.BlsKeysObjectKey(addr))
	return nil
}

func (bk BlsKeysState) Exists(addr string) bool {
	blsKeysKey := types.BlsKeysObjectKey(addr)
	return bk.blsKeys.Has(blsKeysKey)
}
