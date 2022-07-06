package keeper

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RegistrationState struct {
	cdc codec.BinaryCodec
	// addrToBlsKeys maps validator addresses to BLS public keys
	addrToBlsKeys sdk.KVStore
	// blsKeysToAddr maps BLS public keys to validator addresses
	blsKeysToAddr sdk.KVStore
}

func (k Keeper) RegistrationState(ctx sdk.Context) RegistrationState {
	// Build the RegistrationState storage
	store := ctx.KVStore(k.storeKey)
	return RegistrationState{
		cdc:           k.cdc,
		addrToBlsKeys: prefix.NewStore(store, types.AddrToBlsKeyPrefix),
		blsKeysToAddr: prefix.NewStore(store, types.BlsKeyToAddrPrefix),
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
	bkToAddrKey := types.BlsKeyToAddrKey(key)
	rawAddr := rs.blsKeysToAddr.Get(bkToAddrKey)
	if rawAddr != nil && types.BytesToValAddr(rawAddr) != valAddr {
		return types.ErrBlsKeyAlreadyExist.Wrapf("same BLS public key is registered by another validator")
	}

	// save concrete BLS public key object
	blsPkKey := types.AddrToBlsKeyKey(valAddr)
	rs.addrToBlsKeys.Set(blsPkKey, key)
	rs.blsKeysToAddr.Set(bkToAddrKey, types.ValAddrToBytes(valAddr))

	return nil
}

// GetBlsPubKey retrieves BLS public key by validator's address
func (rs RegistrationState) GetBlsPubKey(addr types.ValidatorAddress) (bls12381.PublicKey, error) {
	pkKey := types.AddrToBlsKeyKey(addr)
	rawBytes := rs.addrToBlsKeys.Get(pkKey)
	if rawBytes == nil {
		return nil, types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}
	pk := new(bls12381.PublicKey)
	err := pk.Unmarshal(rawBytes)

	return *pk, err
}

// Exists checks whether a BLS key exists
func (rs RegistrationState) Exists(addr types.ValidatorAddress) bool {
	pkKey := types.AddrToBlsKeyKey(addr)
	return rs.addrToBlsKeys.Has(pkKey)
}
