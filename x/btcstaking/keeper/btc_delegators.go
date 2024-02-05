package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// AddBTCDelegation indexes the given BTC delegation in the BTC delegator store, and saves
// it under BTC delegation store
func (k Keeper) AddBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) error {
	if err := btcDel.ValidateBasic(); err != nil {
		return err
	}

	// for each finality provider the delegation restakes to, update its index
	if err := k.indexBTCDelegation(ctx, btcDel); err != nil {
		return err
	}

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	return nil
}

// updateBTCDelegation updates an existing BTC delegation w.r.t. finality provider BTC PK, delegator BTC PK,
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

// AddCovenantSigsToBTCDelegation adds covenant signatures to a BTC delegation
// with the given staking tx hash, including
// - a list of adaptor signatures over slashing tx, each encrypted by a restaked finality provider's PK
// - a Schnorr signature over unbonding tx
// - a list of adaptor signatures over unbonding slashing tx, each encrypted by a restaked finality provider's PK
func (k Keeper) AddCovenantSigsToBTCDelegation(
	ctx context.Context,
	stakingTxHash string,
	covPk *bbn.BIP340PubKey,
	slashingSigsByte [][]byte,
	unbondingTxSigInfo *bbn.BIP340Signature,
	slashUnbondingTxSigsByte [][]byte,
) error {
	quorum := k.GetParams(ctx).CovenantQuorum

	slashingSigs := make([]asig.AdaptorSignature, 0, len(slashingSigsByte))
	for _, s := range slashingSigsByte {
		as, err := asig.NewAdaptorSignatureFromBytes(s)
		if err != nil {
			return err
		}
		slashingSigs = append(slashingSigs, *as)
	}
	slashUnbondingTxSigs := make([]asig.AdaptorSignature, 0, len(slashUnbondingTxSigsByte))
	for _, s := range slashUnbondingTxSigsByte {
		as, err := asig.NewAdaptorSignatureFromBytes(s)
		if err != nil {
			return err
		}
		slashUnbondingTxSigs = append(slashUnbondingTxSigs, *as)
	}

	addCovenantSig := func(btcDel *types.BTCDelegation) error {
		if err := btcDel.AddCovenantSigs(covPk, slashingSigs, quorum); err != nil {
			return err
		}
		if err := btcDel.BtcUndelegation.AddCovenantSigs(covPk, unbondingTxSigInfo, slashUnbondingTxSigs, quorum); err != nil {
			return err
		}
		// index the newly active BTC delegation
		// TODO: this is not the ideal place for indexing active BTC delegation
		k.indexActiveBTCDelegation(ctx, btcDel)
		return nil
	}

	return k.updateBTCDelegation(ctx, stakingTxHash, addCovenantSig)
}

// hasBTCDelegatorDelegations checks if the given BTC delegator has any BTC delegations under a given finality provider
func (k Keeper) hasBTCDelegatorDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) bool {
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	if !k.HasFinalityProvider(ctx, fpBTCPKBytes) {
		return false
	}
	store := k.btcDelegatorStore(ctx, fpBTCPK)
	return store.Has(delBTCPKBytes)
}

// indexBTCDelegation adds the BTC delegation to the BTC delegator delegation index
func (k Keeper) indexBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) error {
	var err error
	stakingTxHash := btcDel.MustGetStakingTxHash()

	// for each finality provider the delegation restakes to, update its index
	for _, fpBTCPK := range btcDel.FpBtcPkList {
		var btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		if k.hasBTCDelegatorDelegations(ctx, &fpBTCPK, btcDel.BtcPk) {
			btcDelIndex, err = k.getBTCDelegatorDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk)
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
		store := k.btcDelegatorStore(ctx, &fpBTCPK)
		delBTCPKBytes := btcDel.BtcPk.MustMarshal()
		btcDelIndexBytes := k.cdc.MustMarshal(btcDelIndex)
		store.Set(delBTCPKBytes, btcDelIndexBytes)
	}

	return nil
}

// getBTCDelegatorDelegationIndex gets the BTC delegation index with a given BTC PK under a given finality provider
func (k Keeper) getBTCDelegatorDelegationIndex(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegationIndex, error) {
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()
	store := k.btcDelegatorStore(ctx, fpBTCPK)

	// ensure the finality provider exists
	if !k.HasFinalityProvider(ctx, fpBTCPKBytes) {
		return nil, types.ErrFpNotFound
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

// getBTCDelegatorDelegations gets the BTC delegations with a given BTC PK under a given finality provider
func (k Keeper) getBTCDelegatorDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegations, error) {
	btcDelIndex, err := k.getBTCDelegatorDelegationIndex(ctx, fpBTCPK, delBTCPK)
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
// prefix: BTCDelegatorKey || finality provider's Bitcoin secp256k1 PK
// key: delegator's Bitcoin secp256k1 PK
// value: BTCDelegatorDelegationIndex (a list of BTCDelegations' staking tx hashes)
func (k Keeper) btcDelegatorStore(ctx context.Context, fpBTCPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	delegationStore := prefix.NewStore(storeAdapter, types.BTCDelegatorKey)
	return prefix.NewStore(delegationStore, fpBTCPK.MustMarshal())
}
