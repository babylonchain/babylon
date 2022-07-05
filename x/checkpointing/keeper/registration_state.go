package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type RegistrationState struct {
	cdc                 codec.BinaryCodec
	blsKeys             sdk.KVStore
	msgCreateValidators sdk.KVStore
}

func (k Keeper) RegistrationState(ctx sdk.Context) RegistrationState {
	// Build the RegistrationState storage
	store := ctx.KVStore(k.storeKey)
	return RegistrationState{
		cdc:                 k.cdc,
		blsKeys:             prefix.NewStore(store, types.BlsKeysObjectPrefix),
		msgCreateValidators: prefix.NewStore(store, types.MsgCreateValidatorsPrefix),
	}
}

// CreateRegistration inserts the BLS key as well as a corresponding MsgCreateValidator message into the storage
func (rs RegistrationState) CreateRegistration(key *types.BlsPubKey, msg *stakingtypes.MsgCreateValidator) error {
	if rs.Exists(key.Address) {
		return types.ErrBlsKeyAlreadyExist.Wrapf("existed public key: %x", key.Key)
	}

	blsKeysKey := types.BlsKeysObjectKey(key.Address)
	msgKey := types.MsgCreateValidatorsKey(key.Address)

	// save concrete BLS public key object and msgCreateValidator
	rs.blsKeys.Set(blsKeysKey, types.BlsPubKeyToBytes(rs.cdc, key))
	rs.msgCreateValidators.Set(msgKey, rs.cdc.MustMarshal(msg))

	return nil
}

// GetBlsPubKey retrieves BLS public key by validator's address
func (rs RegistrationState) GetBlsPubKey(addr string) (*types.BlsPubKey, error) {
	pubKeyKey := types.BlsKeysObjectKey(addr)
	rawBytes := rs.blsKeys.Get(pubKeyKey)
	if rawBytes == nil {
		return nil, types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}

	return types.BytesToBlsPubKey(rs.cdc, rawBytes)
}

// RemoveBlsPubKey removes a BLS public key
func (rs RegistrationState) RemoveBlsPubKey(addr string) error {
	if !rs.Exists(addr) {
		return types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}

	// delete BLS public key and corresponding msgCreateValidator from storage
	rs.blsKeys.Delete(types.BlsKeysObjectKey(addr))
	rs.msgCreateValidators.Delete(types.MsgCreateValidatorsKey(addr))

	return nil
}

// RemoveMsgCreateValidator removes a MsgCreateValidator
func (rs RegistrationState) RemoveMsgCreateValidator(addr string) {
	rs.msgCreateValidators.Delete(types.MsgCreateValidatorsKey(addr))
}

// Exists checks whether a BLS key exists
func (rs RegistrationState) Exists(addr string) bool {
	blsKeysKey := types.BlsKeysObjectKey(addr)
	return rs.blsKeys.Has(blsKeysKey)
}
