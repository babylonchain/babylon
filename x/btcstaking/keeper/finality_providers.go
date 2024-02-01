package keeper

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// SetFinalityProvider adds the given finality provider to KVStore
func (k Keeper) SetFinalityProvider(ctx context.Context, fp *types.FinalityProvider) {
	store := k.finalityProviderStore(ctx)
	fpBytes := k.cdc.MustMarshal(fp)
	store.Set(fp.BtcPk.MustMarshal(), fpBytes)
}

// HasFinalityProvider checks if the finality provider exists
func (k Keeper) HasFinalityProvider(ctx context.Context, fpBTCPK []byte) bool {
	store := k.finalityProviderStore(ctx)
	return store.Has(fpBTCPK)
}

// GetFinalityProvider gets the finality provider with the given finality provider Bitcoin PK
func (k Keeper) GetFinalityProvider(ctx context.Context, fpBTCPK []byte) (*types.FinalityProvider, error) {
	store := k.finalityProviderStore(ctx)
	if !k.HasFinalityProvider(ctx, fpBTCPK) {
		return nil, types.ErrFpNotFound
	}
	fpBytes := store.Get(fpBTCPK)
	var fp types.FinalityProvider
	k.cdc.MustUnmarshal(fpBytes, &fp)
	return &fp, nil
}

// SlashFinalityProvider slashes a finality provider with the given PK
// A slashed finality provider will not have voting power
func (k Keeper) SlashFinalityProvider(ctx context.Context, fpBTCPK []byte) error {
	fp, err := k.GetFinalityProvider(ctx, fpBTCPK)
	if err != nil {
		return err
	}
	if fp.IsSlashed() {
		return types.ErrFpAlreadySlashed
	}
	fp.SlashedBabylonHeight = uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		panic(fmt.Errorf("failed to get current BTC tip"))
	}
	fp.SlashedBtcHeight = btcTip.Height
	k.SetFinalityProvider(ctx, fp)

	// record newly slashed finality provider and BTC delegations
	types.RecordNewSlashedFinalityProvider()
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcDelIter := k.btcDelegatorStore(ctx, fp.BtcPk).Iterator(nil, nil)
	for ; btcDelIter.Valid(); btcDelIter.Next() {
		// unmarshal
		var btcDelIndex types.BTCDelegatorDelegationIndex
		k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDelIndex)
		// retrieve and process each of the BTC delegation
		for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
			stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
			if err != nil {
				panic(err) // only programming error is possible
			}
			btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
			if btcDel.GetStatus(btcTip.Height, wValue, covenantQuorum) == types.BTCDelegationStatus_ACTIVE {
				types.RecordNewSlashedBTCDelegation()
			}
		}
	}
	btcDelIter.Close()

	return nil
}

// finalityProviderStore returns the KVStore of the finality provider set
// prefix: FinalityProviderKey
// key: Bitcoin secp256k1 PK
// value: FinalityProvider object
func (k Keeper) finalityProviderStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.FinalityProviderKey)
}
