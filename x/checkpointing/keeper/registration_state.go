package keeper

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RegistrationState struct {
	cdc     codec.BinaryCodec
	blsKeys sdk.KVStore
	// keySet maps BLS public keys to validator addresses
	keySet sdk.KVStore
}

func (k Keeper) RegistrationState(ctx sdk.Context) RegistrationState {
	// Build the RegistrationState storage
	store := ctx.KVStore(k.storeKey)
	return RegistrationState{
		cdc:     k.cdc,
		blsKeys: prefix.NewStore(store, types.BlsKeysObjectPrefix),
		keySet:  prefix.NewStore(store, types.BlsKeySetPrefix),
	}
}

// CreateRegistration inserts the BLS key into the addr -> key and key -> addr storage
func (rs RegistrationState) CreateRegistration(key bls12381.PublicKey, valAddr types.ValidatorAddress) error {
	blsPubKey, err := rs.GetBlsPubKey(valAddr)

	// we should disallow a validator to register with different BLS public keys
	if err == nil && !blsPubKey.Equal(key) {
		return types.ErrBlsKeyAlreadyExist.Wrapf("the validator has registered a BLS public key")
	}

	// we should disallow the same BLS public key is registered by different validators
	blsKeySetKey := types.BlsKeySetKey(key)
	rawAddr := rs.keySet.Get(blsKeySetKey)
	if rawAddr != nil && types.BytesToValAddr(rawAddr) != valAddr {
		return types.ErrBlsKeyAlreadyExist.Wrapf("same BLS public key is registered by another validator")
	}

	// save concrete BLS public key object and msgCreateValidator
	blsKeysKey := types.BlsKeysObjectKey(valAddr)
	rs.blsKeys.Set(blsKeysKey, key)
	rs.keySet.Set(blsKeySetKey, types.ValAddrToBytes(valAddr))

	return nil
}

// GetBlsPubKey retrieves BLS public key by validator's address
func (rs RegistrationState) GetBlsPubKey(addr types.ValidatorAddress) (bls12381.PublicKey, error) {
	pubKeyKey := types.BlsKeysObjectKey(addr)
	rawBytes := rs.blsKeys.Get(pubKeyKey)
	if rawBytes == nil {
		return nil, types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}
	pk := new(bls12381.PublicKey)
	err := pk.Unmarshal(rawBytes)

	return *pk, err
}

// RemoveBlsKey removes a BLS public key
// this should be called when a validator is removed
func (rs RegistrationState) RemoveBlsKey(addr types.ValidatorAddress) error {
	blsPubKey, err := rs.GetBlsPubKey(addr)
	if err != nil {
		return types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}

	// delete BLS public key and corresponding key set from storage
	rs.blsKeys.Delete(types.BlsKeysObjectKey(addr))
	rs.keySet.Delete(types.BlsKeySetKey(blsPubKey))

	return nil
}

// Exists checks whether a BLS key exists
func (rs RegistrationState) Exists(addr types.ValidatorAddress) bool {
	blsKeysKey := types.BlsKeysObjectKey(addr)
	return rs.blsKeys.Has(blsKeysKey)
}
