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
func (k Keeper) AddBTCDelegation(ctx sdk.Context, btcDel *types.BTCDelegation) error {
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
		// get BTC delegation index under this finality provider
		btcDelIndex := k.getBTCDelegatorDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk)
		if btcDelIndex == nil {
			btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		}
		// index staking tx hash of this BTC delegation
		if err := btcDelIndex.Add(stakingTxHash); err != nil {
			return types.ErrInvalidStakingTx.Wrapf(err.Error())
		}
		// save the index
		k.setBTCDelegatorDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk, btcDelIndex)
	}

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	// notify subscriber
	event := &types.EventBTCDelegationStateUpdate{
		StakingTxHash: stakingTxHash.String(),
		NewState:      types.BTCDelegationStatus_PENDING,
	}
	if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventBTCDelegationStateUpdate for the new pending BTC delegation: %w", err))
	}

	// NOTE: we don't need to record events for pending BTC delegations since these
	// do not affect voting power distribution

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
	ctx sdk.Context,
	btcDel *types.BTCDelegation,
	covPK *bbn.BIP340PubKey,
	parsedSlashingAdaptorSignatures []asig.AdaptorSignature,
	unbondingTxSig *bbn.BIP340Signature,
	parsedUnbondingSlashingAdaptorSignatures []asig.AdaptorSignature,
	params *types.Params,
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
	if len(btcDel.CovenantSigs) == int(params.CovenantQuorum) {
		// notify subscriber
		event := &types.EventBTCDelegationStateUpdate{
			StakingTxHash: btcDel.MustGetStakingTxHash().String(),
			NewState:      types.BTCDelegationStatus_ACTIVE,
		}
		if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
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
	ctx sdk.Context,
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

	if err := ctx.EventManager().EmitTypedEvent(event); err != nil {
		panic(fmt.Errorf("failed to emit EventBTCDelegationStateUpdate for the new unbonded BTC delegation: %w", err))
	}

	// record event that the BTC delegation becomes unbonded at this height
	unbondedEvent := types.NewEventPowerDistUpdateWithBTCDel(event)
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	k.addPowerDistUpdateEvent(ctx, btcTip.Height, unbondedEvent)
}

func (k Keeper) setBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) {
	store := k.btcDelegationStore(ctx)
	stakingTxHash := btcDel.MustGetStakingTxHash()
	btcDelBytes := k.cdc.MustMarshal(btcDel)
	store.Set(stakingTxHash[:], btcDelBytes)
}

// GetBTCDelegation gets the BTC delegation with a given staking tx hash
func (k Keeper) GetBTCDelegation(ctx context.Context, stakingTxHashStr string) (*types.BTCDelegation, error) {
	// decode staking tx hash string
	stakingTxHash, err := chainhash.NewHashFromStr(stakingTxHashStr)
	if err != nil {
		return nil, err
	}
	btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
	if btcDel == nil {
		return nil, types.ErrBTCDelegationNotFound
	}

	return btcDel, nil
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
