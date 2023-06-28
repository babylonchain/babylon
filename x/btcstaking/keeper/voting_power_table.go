package keeper

import (
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RecordVotingPowerTable computes the voting power table at the current block height
// and saves the power table to KVStore
// triggered upon each EndBlock
func (k Keeper) RecordVotingPowerTable(ctx sdk.Context) {
	// tip of Babylon and Bitcoin
	babylonTipHeight := uint64(ctx.BlockHeight())
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		return
	}
	btcTipHeight := btcTip.Height
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// iterate all BTC validators
	btcValIter := k.btcValidatorStore(ctx).Iterator(nil, nil)
	defer btcValIter.Close()
	for ; btcValIter.Valid(); btcValIter.Next() {
		valBTCPK := btcValIter.Key()
		valPower := uint64(0)

		// iterate all BTC delegations under this validator
		// to calculate this validator's total voting power
		btcDelIter := k.btcDelegationStore(ctx, valBTCPK).Iterator(nil, nil)
		for ; btcDelIter.Valid(); btcDelIter.Next() {
			var btcDel types.BTCDelegation
			k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDel)
			valPower += btcDel.VotingPower(btcTipHeight, wValue)
		}
		btcDelIter.Close()

		if valPower > 0 {
			k.setVotingPower(ctx, valBTCPK, babylonTipHeight, valPower)
		}
	}
}

// setVotingPower sets the voting power of a given BTC validator at a given Babylon height
func (k Keeper) setVotingPower(ctx sdk.Context, valBTCPK []byte, height uint64, power uint64) {
	store := k.votingPowerStore(ctx, height)
	store.Set(valBTCPK, sdk.Uint64ToBigEndian(power))
}

// GetVotingPower gets the voting power of a given BTC validator at a given Babylon height
func (k Keeper) GetVotingPower(ctx sdk.Context, valBTCPK []byte, height uint64) uint64 {
	if !k.HasBTCValidator(ctx, valBTCPK) {
		return 0
	}
	store := k.votingPowerStore(ctx, height)
	powerBytes := store.Get(valBTCPK)
	if len(powerBytes) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(powerBytes)
}

// votingPowerStore returns the KVStore of the BTC validators' voting power
// prefix: (VotingPowerKey || Babylon block height)
// key: Bitcoin secp256k1 PK
// value: voting power quantified in Satoshi
func (k Keeper) votingPowerStore(ctx sdk.Context, height uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	votingPowerStore := prefix.NewStore(store, types.VotingPowerKey)
	return prefix.NewStore(votingPowerStore, sdk.Uint64ToBigEndian(height))
}
