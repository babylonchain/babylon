package keeper

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IterateActiveFPsAndBTCDelegations iterates over all finality providers that are not slashed,
// and their BTC delegations
func (k Keeper) IterateActiveFPsAndBTCDelegations(ctx context.Context, handler func(fp *types.FinalityProvider, btcDel *types.BTCDelegation)) {
	// filter out all finality providers with positive voting power
	fpIter := k.finalityProviderStore(ctx).Iterator(nil, nil)
	defer fpIter.Close()
	for ; fpIter.Valid(); fpIter.Next() {
		fpBTCPKBytes := fpIter.Key()
		fpBTCPK, err := bbn.NewBIP340PubKey(fpBTCPKBytes)
		if err != nil {
			// failed to unmarshal finality provider PK in KVStore is a programming error
			panic(err)
		}
		fp, err := k.GetFinalityProvider(ctx, fpBTCPKBytes)
		if err != nil {
			// failed to get a finality provider with voting power is a programming error
			panic(err)
		}
		if fp.IsSlashed() {
			// slashed finality provider is removed from finality provider set
			continue
		}

		// iterate all BTC delegations under this finality provider
		// to calculate this finality provider's total voting power
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
				handler(fp, btcDel)
			}
		}
	}
}

// setCurrentTopNVotingPower gets top N finality providers and set their current voting power to KV store
func (k Keeper) setCurrentTopNVotingPower(ctx context.Context, fpPowerMap map[string]uint64) {
	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	activeFps := []*types.FinalityProviderWithMeta{}
	for btcPKHex, power := range fpPowerMap {
		btcPK, err := bbn.NewBIP340PubKeyFromHex(btcPKHex)
		if err != nil {
			panic(err) // only programming error
		}
		activeFps = append(activeFps, &types.FinalityProviderWithMeta{BtcPk: btcPK, VotingPower: power})
	}
	activeFps = types.FilterTopNFinalityProviders(activeFps, k.GetParams(ctx).MaxActiveFinalityProviders)

	// get current Babylon height
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	// set voting power for each active finality providers
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.VotingPower)
	}
}

// SetVotingPower sets the voting power of a given finality provider at a given Babylon height
func (k Keeper) SetVotingPower(ctx context.Context, fpBTCPK []byte, height uint64, power uint64) {
	store := k.votingPowerStore(ctx, height)
	store.Set(fpBTCPK, sdk.Uint64ToBigEndian(power))
}

// GetVotingPower gets the voting power of a given finality provider at a given Babylon height
func (k Keeper) GetVotingPower(ctx context.Context, fpBTCPK []byte, height uint64) uint64 {
	if !k.HasFinalityProvider(ctx, fpBTCPK) {
		return 0
	}
	store := k.votingPowerStore(ctx, height)
	powerBytes := store.Get(fpBTCPK)
	if len(powerBytes) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(powerBytes)
}

// GetCurrentVotingPower gets the voting power of a given finality provider at the current height
// NOTE: it's possible that the voting power table is 1 block behind CometBFT, e.g., when `BeginBlock`
// hasn't executed yet
func (k Keeper) GetCurrentVotingPower(ctx context.Context, fpBTCPK []byte) (uint64, uint64) {
	// find the last recorded voting power table via iterator
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.VotingPowerKey)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	// no voting power table is known yet, return 0
	if !iter.Valid() {
		return 0, 0
	}

	// there is known voting power table, find the last height
	lastHeight := sdk.BigEndianToUint64(iter.Key())
	storeAtHeight := prefix.NewStore(store, sdk.Uint64ToBigEndian(lastHeight))

	// if the finality provider is not known, return 0 voting power
	if !k.HasFinalityProvider(ctx, fpBTCPK) {
		return lastHeight, 0
	}

	// find the voting power of this finality provider
	powerBytes := storeAtHeight.Get(fpBTCPK)
	if len(powerBytes) == 0 {
		return lastHeight, 0
	}

	return lastHeight, sdk.BigEndianToUint64(powerBytes)
}

// HasVotingPowerTable checks if the voting power table exists at a given height
func (k Keeper) HasVotingPowerTable(ctx context.Context, height uint64) bool {
	store := k.votingPowerStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	return iter.Valid()
}

// GetVotingPowerTable gets the voting power table, i.e., finality provider set at a given height
func (k Keeper) GetVotingPowerTable(ctx context.Context, height uint64) map[string]uint64 {
	store := k.votingPowerStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// if no finality provider at this height, return nil
	if !iter.Valid() {
		return nil
	}

	// get all finality providers at this height
	fpSet := map[string]uint64{}
	for ; iter.Valid(); iter.Next() {
		fpBTCPK, err := bbn.NewBIP340PubKey(iter.Key())
		if err != nil {
			// failing to unmarshal finality provider BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}
		fpSet[fpBTCPK.MarshalHex()] = sdk.BigEndianToUint64(iter.Value())
	}

	return fpSet
}

// GetBTCStakingActivatedHeight returns the height when the BTC staking protocol is activated
// i.e., the first height where a finality provider has voting power
// Before the BTC staking protocol is activated, we don't index or tally any block
func (k Keeper) GetBTCStakingActivatedHeight(ctx context.Context) (uint64, error) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	votingPowerStore := prefix.NewStore(storeAdapter, types.VotingPowerKey)
	iter := votingPowerStore.Iterator(nil, nil)
	defer iter.Close()
	// if the iterator is valid, then there exists a height that has a finality provider with voting power
	if iter.Valid() {
		return sdk.BigEndianToUint64(iter.Key()), nil
	} else {
		return 0, types.ErrBTCStakingNotActivated
	}
}

func (k Keeper) IsBTCStakingActivated(ctx context.Context) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	votingPowerStore := prefix.NewStore(storeAdapter, types.VotingPowerKey)
	iter := votingPowerStore.Iterator(nil, nil)
	defer iter.Close()
	// if the iterator is valid, then BTC staking is already activated
	return iter.Valid()
}

// votingPowerStore returns the KVStore of the finality providers' voting power
// prefix: (VotingPowerKey || Babylon block height)
// key: Bitcoin secp256k1 PK
// value: voting power quantified in Satoshi
func (k Keeper) votingPowerStore(ctx context.Context, height uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	votingPowerStore := prefix.NewStore(storeAdapter, types.VotingPowerKey)
	return prefix.NewStore(votingPowerStore, sdk.Uint64ToBigEndian(height))
}
