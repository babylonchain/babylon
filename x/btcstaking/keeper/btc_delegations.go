package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetBTCDelegation(ctx sdk.Context, btcDel *types.BTCDelegation) error {
	var (
		btcDels = types.NewBTCDelegatorDelegations()
		err     error
	)
	if k.hasBTCDelegations(ctx, btcDel.ValBtcPk, btcDel.BtcPk) {
		btcDels, err = k.getBTCDelegations(ctx, btcDel.ValBtcPk, btcDel.BtcPk)
		if err != nil {
			// this can only be a programming error
			panic(fmt.Errorf("failed to get BTC delegations while hasBTCDelegations returns true"))
		}
	}
	if err := btcDels.Add(btcDel); err != nil {
		return types.ErrInvalidStakingTx.Wrapf(err.Error())
	}

	k.setBTCDelegations(ctx, btcDel.ValBtcPk, btcDel.BtcPk, btcDels)
	return nil
}

// AddJurySigToBTCDelegation adds a given jury sig to a BTC delegation
// with the given (val PK, del PK, staking tx hash) tuple
func (k Keeper) AddJurySigToBTCDelegation(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, stakingTxHash string, jurySig *bbn.BIP340Signature) error {
	btcDels, err := k.getBTCDelegations(ctx, valBTCPK, delBTCPK)
	if err != nil {
		return err
	}
	if err := btcDels.AddJurySig(stakingTxHash, jurySig); err != nil {
		return types.ErrInvalidJurySig.Wrapf(err.Error())
	}
	k.setBTCDelegations(ctx, valBTCPK, delBTCPK, btcDels)
	return nil
}

// setBTCDelegations sets the given BTC delegation to KVStore
func (k Keeper) setBTCDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, btcDels *types.BTCDelegatorDelegations) {
	delBTCPKBytes := delBTCPK.MustMarshal()

	store := k.btcDelegationStore(ctx, valBTCPK)
	btcDelBytes := k.cdc.MustMarshal(btcDels)
	store.Set(delBTCPKBytes, btcDelBytes)
}

// hasBTCDelegations checks if the given BTC delegator has any BTC delegations under a given BTC validator
func (k Keeper) hasBTCDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) bool {
	valBTCPKBytes := valBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	if !k.HasBTCValidator(ctx, valBTCPKBytes) {
		return false
	}
	store := k.btcDelegationStore(ctx, valBTCPK)
	return store.Has(delBTCPKBytes)
}

// validatorDelegations gets the BTC delegations with a given BTC PK under a given BTC validator
// NOTE: Internal function which assumes that the validator exists
func (k Keeper) validatorDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPKBytes []byte) (*types.BTCDelegatorDelegations, error) {
	store := k.btcDelegationStore(ctx, valBTCPK)
	// ensure BTC delegation exists
	if !store.Has(delBTCPKBytes) {
		return nil, types.ErrBTCDelNotFound
	}
	// get and unmarshal
	var btcDels types.BTCDelegatorDelegations
	btcDelsBytes := store.Get(delBTCPKBytes)
	k.cdc.MustUnmarshal(btcDelsBytes, &btcDels)
	return &btcDels, nil
}

// getBTCDelegations gets the BTC delegations with a given BTC PK under a given BTC validator
func (k Keeper) getBTCDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegations, error) {
	valBTCPKBytes := valBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	// ensure the BTC validator exists
	if !k.HasBTCValidator(ctx, valBTCPKBytes) {
		return nil, types.ErrBTCValNotFound
	}

	return k.validatorDelegations(ctx, valBTCPK, delBTCPKBytes)
}

// GetBTCDelegation gets the BTC delegation with a given BTC PK and staking tx hash under a given BTC validator
func (k Keeper) GetBTCDelegation(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, stakingTxHash string) (*types.BTCDelegation, error) {
	btcDels, err := k.getBTCDelegations(ctx, valBTCPK, delBTCPK)
	if err != nil {
		return nil, err
	}
	btcDel, err := btcDels.Get(stakingTxHash)
	if err != nil {
		return nil, types.ErrBTCDelNotFound.Wrapf(err.Error())
	}
	return btcDel, nil
}

// btcDelegationStore returns the KVStore of the BTC delegations
// prefix: BTCDelegationKey || validator's Bitcoin secp256k1 PK
// key: delegation's Bitcoin secp256k1 PK
// value: BTCDelegations (a list of BTCDelegation)
func (k Keeper) btcDelegationStore(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.BTCDelegationKey)
	return prefix.NewStore(delegationStore, valBTCPK.MustMarshal())
}
