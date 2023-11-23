package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AddBTCDelegation indexes the given BTC delegation in the BTC delegator store, and saves
// it under BTC delegation store
func (k Keeper) AddBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) error {
	if err := btcDel.ValidateBasic(); err != nil {
		return err
	}

	// get staking tx hash
	stakingTxHash, err := btcDel.GetStakingTxHash()
	if err != nil {
		return err
	}

	// for each BTC validator the delegation restakes to, update its index
	for _, valBTCPK := range btcDel.ValBtcPkList {
		var btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		if k.hasBTCDelegatorDelegations(ctx, &valBTCPK, btcDel.BtcPk) {
			btcDelIndex, err = k.getBTCDelegatorDelegationIndex(ctx, &valBTCPK, btcDel.BtcPk)
			if err != nil {
				// this can only be a programming error
				panic(fmt.Errorf("failed to get BTC delegations while hasBTCDelegatorDelegations returns true"))
			}
		}

		// index staking tx hash of this BTC delegation
		if err := btcDelIndex.Add(stakingTxHash); err != nil {
			return types.ErrInvalidStakingTx.Wrapf(err.Error())
		}
		// save the index
		store := k.btcDelegatorStore(ctx, &valBTCPK)
		delBTCPKBytes := btcDel.BtcPk.MustMarshal()
		btcDelIndexBytes := k.cdc.MustMarshal(btcDelIndex)
		store.Set(delBTCPKBytes, btcDelIndexBytes)
	}

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	return nil
}

// updateBTCDelegation updates an existing BTC delegation w.r.t. validator BTC PK, delegator BTC PK,
// and staking tx hash by using a given function
func (k Keeper) updateBTCDelegation(
	ctx context.Context,
	stakingTxHashStr string,
	modifyFn func(*types.BTCDelegation) error,
) error {
	// get the BTC delegation
	btcDel, err := k.GetBTCDelegation(ctx, stakingTxHashStr)
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

// AddCovenantSigToBTCDelegation adds a given covenant sig to a BTC delegation
// with the given (val PK, del PK, staking tx hash) tuple
func (k Keeper) AddCovenantSigToBTCDelegation(
	ctx context.Context,
	stakingTxHash string,
	covenantSig *bbn.BIP340Signature,
) error {
	addCovenantSig := func(btcDel *types.BTCDelegation) error {
		if btcDel.CovenantSig != nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s already has a covenant signature", stakingTxHash)
		}
		btcDel.CovenantSig = covenantSig
		return nil
	}

	return k.updateBTCDelegation(ctx, stakingTxHash, addCovenantSig)
}

func (k Keeper) AddUndelegationToBTCDelegation(
	ctx context.Context,
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

	return k.updateBTCDelegation(ctx, stakingTxHash, addUndelegation)
}

func (k Keeper) AddCovenantSigsToUndelegation(
	ctx context.Context,
	stakingTxHash string,
	unbondingTxSigInfo *types.SignatureInfo,
	slashUnbondingTxSig *bbn.BIP340Signature,
) error {
	addCovenantSigs := func(btcDel *types.BTCDelegation) error {
		if btcDel.BtcUndelegation == nil {
			return fmt.Errorf("the BTC delegation with staking tx hash %s did not receive undelegation request yet", stakingTxHash)
		}

		// TODO: ensure the given CovenantUnbondingSig does not exist in the BTC undelegation yet
		if btcDel.BtcUndelegation.CovenantUnbondingSigList != nil || btcDel.BtcUndelegation.CovenantSlashingSig != nil {
			return fmt.Errorf("the BTC undelegation for staking tx hash %s already has valid covenant signatures", stakingTxHash)
		}

		btcDel.BtcUndelegation.CovenantUnbondingSigList = append(btcDel.BtcUndelegation.CovenantUnbondingSigList, unbondingTxSigInfo)
		btcDel.BtcUndelegation.CovenantSlashingSig = slashUnbondingTxSig
		return nil
	}

	return k.updateBTCDelegation(ctx, stakingTxHash, addCovenantSigs)
}

// hasBTCDelegatorDelegations checks if the given BTC delegator has any BTC delegations under a given BTC validator
func (k Keeper) hasBTCDelegatorDelegations(ctx context.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) bool {
	valBTCPKBytes := valBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	if !k.HasBTCValidator(ctx, valBTCPKBytes) {
		return false
	}
	store := k.btcDelegatorStore(ctx, valBTCPK)
	return store.Has(delBTCPKBytes)
}

// getBTCDelegatorDelegationIndex gets the BTC delegation index with a given BTC PK under a given BTC validator
func (k Keeper) getBTCDelegatorDelegationIndex(ctx context.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegationIndex, error) {
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
func (k Keeper) getBTCDelegatorDelegations(ctx context.Context, valBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegations, error) {
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

// GetBTCDelegation gets the BTC delegation with a given staking tx hash
func (k Keeper) GetBTCDelegation(ctx context.Context, stakingTxHashStr string) (*types.BTCDelegation, error) {
	// decode staking tx hash string
	stakingTxHash, err := chainhash.NewHashFromStr(stakingTxHashStr)
	if err != nil {
		return nil, err
	}

	return k.getBTCDelegation(ctx, *stakingTxHash), nil
}

// btcDelegatorStore returns the KVStore of the BTC delegators
// prefix: BTCDelegatorKey || validator's Bitcoin secp256k1 PK
// key: delegator's Bitcoin secp256k1 PK
// value: BTCDelegatorDelegationIndex (a list of BTCDelegations' staking tx hashes)
func (k Keeper) btcDelegatorStore(ctx context.Context, valBTCPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	delegationStore := prefix.NewStore(storeAdapter, types.BTCDelegatorKey)
	return prefix.NewStore(delegationStore, valBTCPK.MustMarshal())
}
