package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddBTCDelegation indexes the given BTC delegation in the BTC delegator store, and saves
// it under BTC delegation store
func (k Keeper) AddBTCDelegation(ctx sdk.Context, btcDel *types.BTCDelegation) error {
	var (
		btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		err         error
	)
	if k.hasBTCDelegatorDelegations(ctx, btcDel.ValBtcPk, btcDel.BtcPk) {
		btcDelIndex, err = k.getBTCDelegatorDelegationIndex(ctx, btcDel.ValBtcPk, btcDel.BtcPk)
		if err != nil {
			// this can only be a programming error
			panic(fmt.Errorf("failed to get BTC delegations while hasBTCDelegatorDelegations returns true"))
		}
	}
	// index staking tx hash of this BTC delegation
	stakingTxHash, err := btcDel.GetStakingTxHash()
	if err != nil {
		return err
	}
	if err := btcDelIndex.Add(stakingTxHash); err != nil {
		return types.ErrInvalidStakingTx.Wrapf(err.Error())
	}
	// save the index
	store := k.btcDelegatorStore(ctx, btcDel.ValBtcPk)
	delBTCPKBytes := btcDel.BtcPk.MustMarshal()
	btcDelIndexBytes := k.cdc.MustMarshal(btcDelIndex)
	store.Set(delBTCPKBytes, btcDelIndexBytes)

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	return nil
}

// updateBTCDelegation updates an existing BTC delegation w.r.t. validator BTC PK, delegator BTC PK,
// and staking tx hash by using a given function
func (k Keeper) updateBTCDelegation(
	ctx sdk.Context,
	valBTCPK *bbn.BIP340PubKey,
	delBTCPK *bbn.BIP340PubKey,
	stakingTxHashStr string,
	modifyFn func(*types.BTCDelegation) error,
) error {
	// get the BTC delegation
	btcDel, err := k.GetBTCDelegation(ctx, valBTCPK, delBTCPK, stakingTxHashStr)
	if err != nil {
		return err
	}

	// apply modification
	if err := modifyFn(btcDel); err != nil {
		return err
	}

	// we only need to update the actual BTC delegation object here, without touching
	// the BTC delegation index
	k.setBTCDelegation(ctx, btcDel)
	return nil
}

// AddJurySigToBTCDelegation adds a given jury sig to a BTC delegation
// with the given (val PK, del PK, staking tx hash) tuple
func (k Keeper) AddJurySigToBTCDelegation(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, stakingTxHash string, jurySig *bbn.BIP340Signature) error {
	addJurySig := func(btcDel *types.BTCDelegation) error {
		if btcDel.JurySig != nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s already has a jury signature", stakingTxHash)
		}
		btcDel.JurySig = jurySig
		return nil
	}

	return k.updateBTCDelegation(ctx, valBTCPK, delBTCPK, stakingTxHash, addJurySig)
}

func (k Keeper) AddUndelegationToBTCDelegation(
	ctx sdk.Context,
	valBTCPK *bbn.BIP340PubKey,
	delBTCPK *bbn.BIP340PubKey,
	stakingTxHash string,
	ud *types.BTCUndelegation,
) error {
	addUndelegation := func(btcDel *types.BTCDelegation) error {
		if btcDel.BtcUndelegation != nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s already has valid undelegation object", stakingTxHash)
		}
		btcDel.BtcUndelegation = ud
		return nil
	}

	return k.updateBTCDelegation(ctx, valBTCPK, delBTCPK, stakingTxHash, addUndelegation)
}

func (k Keeper) AddValidatorSigToUndelegation(
	ctx sdk.Context,
	valBTCPK *bbn.BIP340PubKey,
	delBTCPK *bbn.BIP340PubKey,
	stakingTxHash string,
	sig *bbn.BIP340Signature,
) error {
	addValidatorSig := func(btcDel *types.BTCDelegation) error {
		if btcDel.BtcUndelegation == nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s did not receive undelegation request yet", stakingTxHash)
		}

		if btcDel.BtcUndelegation.ValidatorUnbondingSig != nil {
			return fmt.Errorf("the BTC undelegation for staking tx hash %s already has valid validator signature", stakingTxHash)
		}

		btcDel.BtcUndelegation.ValidatorUnbondingSig = sig
		return nil
	}

	return k.updateBTCDelegation(ctx, valBTCPK, delBTCPK, stakingTxHash, addValidatorSig)
}

func (k Keeper) AddJurySigsToUndelegation(
	ctx sdk.Context,
	valBTCPK *bbn.BIP340PubKey,
	delBTCPK *bbn.BIP340PubKey,
	stakingTxHash string,
	unbondingTxSig *bbn.BIP340Signature,
	slashUnbondingTxSig *bbn.BIP340Signature,
) error {
	addJurySigs := func(btcDel *types.BTCDelegation) error {
		if btcDel.BtcUndelegation == nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s did not receive undelegation request yet", stakingTxHash)
		}

		if btcDel.BtcUndelegation.JuryUnbondingSig != nil || btcDel.BtcUndelegation.JurySlashingSig != nil {
			return fmt.Errorf("the BTC undelegation for staking tx hash %s already has valid jury signatures", stakingTxHash)
		}

		btcDel.BtcUndelegation.JuryUnbondingSig = unbondingTxSig
		btcDel.BtcUndelegation.JurySlashingSig = slashUnbondingTxSig
		return nil
	}

	return k.updateBTCDelegation(ctx, valBTCPK, delBTCPK, stakingTxHash, addJurySigs)
}

// hasBTCDelegatorDelegations checks if the given BTC delegator has any BTC delegations under a given BTC validator
func (k Keeper) hasBTCDelegatorDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) bool {
	valBTCPKBytes := valBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	if !k.HasBTCValidator(ctx, valBTCPKBytes) {
		return false
	}
	store := k.btcDelegatorStore(ctx, valBTCPK)
	return store.Has(delBTCPKBytes)
}

// getBTCDelegatorDelegationIndex gets the BTC delegation index with a given BTC PK under a given BTC validator
func (k Keeper) getBTCDelegatorDelegationIndex(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegationIndex, error) {
	valBTCPKBytes := valBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()
	store := k.btcDelegatorStore(ctx, valBTCPK)

	// ensure the BTC validator exists
	if !k.HasBTCValidator(ctx, valBTCPKBytes) {
		return nil, types.ErrBTCValNotFound
	}

	// ensure BTC delegator exists
	if !store.Has(delBTCPKBytes) {
		return nil, types.ErrBTCDelegatorNotFound
	}
	// get and unmarshal
	var btcDelIndex types.BTCDelegatorDelegationIndex
	btcDelIndexBytes := store.Get(delBTCPKBytes)
	k.cdc.MustUnmarshal(btcDelIndexBytes, &btcDelIndex)
	return &btcDelIndex, nil
}

// getBTCDelegatorDelegations gets the BTC delegations with a given BTC PK under a given BTC validator
func (k Keeper) getBTCDelegatorDelegations(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegations, error) {
	btcDelIndex, err := k.getBTCDelegatorDelegationIndex(ctx, valBTCPK, delBTCPK)
	if err != nil {
		return nil, err
	}
	// get BTC delegation from each staking tx hash
	btcDels := []*types.BTCDelegation{}
	for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
		stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
		if err != nil {
			// failing to unmarshal hash bytes in DB's BTC delegation index is a programming error
			panic(err)
		}
		btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
		btcDels = append(btcDels, btcDel)
	}
	return &types.BTCDelegatorDelegations{Dels: btcDels}, nil
}

// GetBTCDelegation gets the BTC delegation with a given BTC PK and staking tx hash under a given BTC validator
// TODO: only take stakingTxHash as input could be enough?
func (k Keeper) GetBTCDelegation(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, stakingTxHashStr string) (*types.BTCDelegation, error) {
	// find the BTC delegation index
	btcDelIndex, err := k.getBTCDelegatorDelegationIndex(ctx, valBTCPK, delBTCPK)
	if err != nil {
		return nil, err
	}
	// decode staking tx hash string
	stakingTxHash, err := chainhash.NewHashFromStr(stakingTxHashStr)
	if err != nil {
		return nil, err
	}
	// ensure the BTC delegation index has this staking tx hash
	if !btcDelIndex.Has(*stakingTxHash) {
		return nil, types.ErrBTCDelegatorNotFound.Wrapf(err.Error())
	}

	return k.getBTCDelegation(ctx, *stakingTxHash), nil
}

// btcDelegatorStore returns the KVStore of the BTC delegators
// prefix: BTCDelegatorKey || validator's Bitcoin secp256k1 PK
// key: delegator's Bitcoin secp256k1 PK
// value: BTCDelegatorDelegationIndex (a list of BTCDelegations' staking tx hashes)
func (k Keeper) btcDelegatorStore(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.BTCDelegatorKey)
	return prefix.NewStore(delegationStore, valBTCPK.MustMarshal())
}
