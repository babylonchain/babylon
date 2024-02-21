package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddBTCDelegation adds a BTC delegation post verification to the system, including
// - indexing the given BTC delegation in the BTC delegator store,
// - saving it under BTC delegation store, and
// - emit events about this BTC delegation.
func (k Keeper) AddBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) error {
	if err := btcDel.ValidateBasic(); err != nil {
		return err
	}

	// get staking tx hash
	stakingTxHash, err := btcDel.GetStakingTxHash()
	if err != nil {
		return err
	}

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

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	// notify subscriber
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventBTCDelegationStateUpdate{
		StakingTxHash: stakingTxHash.String(),
		NewState:      types.BTCDelegationStatus_PENDING,
	}
	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventBTCDelegationStateUpdate for the new pending BTC delegation: %w", err))
	}

	// record event that the BTC delegation becomes pending at this height
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	pendingEvent := types.NewEventPowerDistUpdateWithBTCDel(event)
	k.addPowerDistUpdateEvent(ctx, btcTip.Height, pendingEvent)
	// record event that the BTC delegation will become unbonded at endHeight-w
	unbondedEvent := types.NewEventPowerDistUpdateWithBTCDel(&types.EventBTCDelegationStateUpdate{
		StakingTxHash: stakingTxHash.String(),
		NewState:      types.BTCDelegationStatus_UNBONDED,
	})
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	k.addPowerDistUpdateEvent(ctx, btcDel.EndHeight-wValue, unbondedEvent)

	return nil
}

// addCovenantSigsToBTCDelegation adds signatures from a given covenant member
// to the given BTC delegation
func (k Keeper) addCovenantSigsToBTCDelegation(
	ctx context.Context,
	btcDel *types.BTCDelegation,
	covPK *bbn.BIP340PubKey,
	parsedSlashingAdaptorSignatures []asig.AdaptorSignature,
	unbondingTxSig *bbn.BIP340Signature,
	parsedUnbondingSlashingAdaptorSignatures []asig.AdaptorSignature,
) {
	// All is fine add received signatures to the BTC delegation and BtcUndelegation
	btcDel.AddCovenantSigs(
		covPK,
		parsedSlashingAdaptorSignatures,
		unbondingTxSig,
		parsedUnbondingSlashingAdaptorSignatures,
	)

	k.setBTCDelegation(ctx, btcDel)

	// If reaching the covenant quorum after this msg, the BTC delegation becomes
	// active. Then, record and emit this event
	if len(btcDel.CovenantSigs) == int(k.GetParams(ctx).CovenantQuorum) {
		// notify subscriber
		event := &types.EventBTCDelegationStateUpdate{
			StakingTxHash: btcDel.MustGetStakingTxHash().String(),
			NewState:      types.BTCDelegationStatus_ACTIVE,
		}
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
			panic(fmt.Errorf("failed to emit EventBTCDelegationStateUpdate for the new active BTC delegation: %w", err))
		}

		// record event that the BTC delegation becomes active at this height
		activeEvent := types.NewEventPowerDistUpdateWithBTCDel(event)
		btcTip := k.btclcKeeper.GetTipInfo(ctx)
		k.addPowerDistUpdateEvent(ctx, btcTip.Height, activeEvent)
	}
}

// btcUndelegate adds the signature of the unbonding tx signed by the staker
// to the given BTC delegation
func (k Keeper) btcUndelegate(
	ctx context.Context,
	btcDel *types.BTCDelegation,
	unbondingTxSig *bbn.BIP340Signature,
) {
	btcDel.BtcUndelegation.DelegatorUnbondingSig = unbondingTxSig
	k.setBTCDelegation(ctx, btcDel)

	// notify subscriber about this unbonded BTC delegation
	event := &types.EventBTCDelegationStateUpdate{
		StakingTxHash: btcDel.MustGetStakingTxHash().String(),
		NewState:      types.BTCDelegationStatus_UNBONDED,
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventBTCDelegationStateUpdate for the new unbonded BTC delegation: %w", err))
	}

	// record event that the BTC delegation becomes unbonded at this height
	unbondedEvent := types.NewEventPowerDistUpdateWithBTCDel(event)
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	k.addPowerDistUpdateEvent(ctx, btcTip.Height, unbondedEvent)
}

// IterateBTCDelegations iterates all BTC delegations under a given finality provider
func (k Keeper) IterateBTCDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, handler func(btcDel *types.BTCDelegation) bool) {
	btcDelIter := k.btcDelegatorStore(ctx, fpBTCPK).Iterator(nil, nil)
	defer btcDelIter.Close()
	for ; btcDelIter.Valid(); btcDelIter.Next() {
		// unmarshal delegator's delegation index
		var btcDelIndex types.BTCDelegatorDelegationIndex
		k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDelIndex)
		// retrieve and process each of the BTC delegation
		for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
			stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
			if err != nil {
				panic(err) // only programming error is possible
			}
			btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
			shouldContinue := handler(btcDel)
			if !shouldContinue {
				return
			}
		}
	}
}

func (k Keeper) setBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) {
	store := k.btcDelegationStore(ctx)
	stakingTxHash := btcDel.MustGetStakingTxHash()
	btcDelBytes := k.cdc.MustMarshal(btcDel)
	store.Set(stakingTxHash[:], btcDelBytes)
}

func (k Keeper) getBTCDelegation(ctx context.Context, stakingTxHash chainhash.Hash) *types.BTCDelegation {
	store := k.btcDelegationStore(ctx)
	btcDelBytes := store.Get(stakingTxHash[:])
	if len(btcDelBytes) == 0 {
		return nil
	}
	var btcDel types.BTCDelegation
	k.cdc.MustUnmarshal(btcDelBytes, &btcDel)
	return &btcDel
}

// btcDelegationStore returns the KVStore of the BTC delegations
// prefix: BTCDelegationKey
// key: BTC delegation's staking tx hash
// value: BTCDelegation
func (k Keeper) btcDelegationStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.BTCDelegationKey)
}
