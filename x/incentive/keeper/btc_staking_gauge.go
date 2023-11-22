package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardBTCStaking distributes rewards to BTC validators/delegations at a given height according
// to the reward distribution cache
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/x/distribution/keeper/allocation.go#L12-L64)
func (k Keeper) RewardBTCStaking(ctx context.Context, height uint64, rdc *bstypes.RewardDistCache) {
	gauge := k.GetBTCStakingGauge(ctx, height)
	if gauge == nil {
		// failing to get a reward gauge at previous height is a programming error
		panic("failed to get a reward gauge at previous height")
	}
	// reward each of the BTC validator and its BTC delegations in proportion
	for _, btcVal := range rdc.BtcVals {
		// get coins that will be allocated to the BTC validator and its BTC delegations
		btcValPortion := rdc.GetBTCValPortion(btcVal)
		coinsForBTCValAndDels := gauge.GetCoinsPortion(btcValPortion)
		// reward the BTC validator with commission
		coinsForCommission := types.GetCoinsPortion(coinsForBTCValAndDels, *btcVal.Commission)
		k.accumulateRewardGauge(ctx, types.BTCValidatorType, btcVal.GetAddress(), coinsForCommission)
		// reward the rest of coins to each BTC delegation proportional to its voting power portion
		coinsForBTCDels := coinsForBTCValAndDels.Sub(coinsForCommission...)
		for _, btcDel := range btcVal.BtcDels {
			btcDelPortion := btcVal.GetBTCDelPortion(btcDel)
			coinsForDel := types.GetCoinsPortion(coinsForBTCDels, btcDelPortion)
			k.accumulateRewardGauge(ctx, types.BTCDelegationType, btcDel.GetAddress(), coinsForDel)
		}
	}

	// TODO: handle the change in the gauge due to the truncating operations
}

func (k Keeper) accumulateBTCStakingReward(ctx context.Context, btcStakingReward sdk.Coins) {
	// update BTC staking gauge
	height := uint64(sdk.UnwrapSDKContext(ctx).BlockHeight())
	gauge := types.NewGauge(btcStakingReward...)
	k.SetBTCStakingGauge(ctx, height, gauge)

	// transfer the BTC staking reward from fee collector account to incentive module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, btcStakingReward)
	if err != nil {
		// this can only be programming error and is unrecoverable
		panic(err)
	}
}

func (k Keeper) SetBTCStakingGauge(ctx context.Context, height uint64, gauge *types.Gauge) {
	store := k.btcStakingGaugeStore(ctx)
	gaugeBytes := k.cdc.MustMarshal(gauge)
	store.Set(sdk.Uint64ToBigEndian(height), gaugeBytes)
}

func (k Keeper) GetBTCStakingGauge(ctx context.Context, height uint64) *types.Gauge {
	store := k.btcStakingGaugeStore(ctx)
	gaugeBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if gaugeBytes == nil {
		return nil
	}

	var gauge types.Gauge
	k.cdc.MustUnmarshal(gaugeBytes, &gauge)
	return &gauge
}

// btcStakingGaugeStore returns the KVStore of the gauge of total reward for
// BTC staking at each height
// prefix: BTCStakingGaugeKey
// key: gauge height
// value: gauge of rewards for BTC staking at this height
func (k Keeper) btcStakingGaugeStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.BTCStakingGaugeKey)
}
