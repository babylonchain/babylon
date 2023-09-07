package keeper

import (
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardBTCStaking distributes rewards to BTC validators/delegations at a given height according
// to the reward distribution cache
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/x/distribution/keeper/allocation.go#L12-L64)
func (k Keeper) RewardBTCStaking(ctx sdk.Context, height uint64, rdc *bstypes.RewardDistCache) {
	gauge, err := k.GetBTCStakingGauge(ctx, height)
	if err != nil {
		// failing to get a reward gauge at previous height is a programming error
		panic(err)
	}
	// reward each of the BTC validator and its BTC delegations in proportion
	for _, btcVal := range rdc.BtcVals {
		// get coins that will be allocated to the BTC validator and its BTC delegations
		btcValPortion := rdc.GetBTCValPortion(btcVal)
		coinsForBTCValAndDels := gauge.GetCoinsPortion(btcValPortion)
		// reward the BTC validator with commission
		coinsForCommission := types.GetCoinsPortion(coinsForBTCValAndDels, *btcVal.Commission)
		if coinsForCommission.IsAllPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, btcVal.GetAddress(), coinsForCommission); err != nil {
				// incentive module account is supposed to have enough balance
				panic(err)
			}
		}
		// reward the rest of coins to each BTC delegation proportional to its voting power portion
		coinsForBTCDels := coinsForBTCValAndDels.Sub(coinsForCommission...)
		for _, btcDel := range btcVal.BtcDels {
			btcDelPortion := btcVal.GetBTCDelPortion(btcDel)
			coinsForDel := types.GetCoinsPortion(coinsForBTCDels, btcDelPortion)
			if coinsForDel.IsAllPositive() {
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, btcDel.GetAddress(), coinsForDel); err != nil {
					// incentive module account is supposed to have enough balance
					panic(err)
				}
			}
		}
	}

	// TODO: handle the change in the gauge due to the truncating operations
}

func (k Keeper) accumulateBTCStakingReward(ctx sdk.Context, btcStakingReward sdk.Coins) {
	// update BTC staking gauge
	height := uint64(ctx.BlockHeight())
	gauge := types.NewGauge(btcStakingReward)
	k.SetBTCStakingGauge(ctx, height, gauge)

	// transfer the BTC staking reward from fee collector account to incentive module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, btcStakingReward)
	if err != nil {
		// this can only be programming error and is unrecoverable
		panic(err)
	}
}

func (k Keeper) SetBTCStakingGauge(ctx sdk.Context, height uint64, gauge *types.Gauge) {
	store := k.btcStakingGaugeStore(ctx)
	gaugeBytes := k.cdc.MustMarshal(gauge)
	store.Set(sdk.Uint64ToBigEndian(height), gaugeBytes)
}

func (k Keeper) GetBTCStakingGauge(ctx sdk.Context, height uint64) (*types.Gauge, error) {
	store := k.btcStakingGaugeStore(ctx)
	gaugeBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(gaugeBytes) == 0 {
		return nil, types.ErrBTCStakingGaugeNotFound
	}

	var gauge types.Gauge
	k.cdc.MustUnmarshal(gaugeBytes, &gauge)
	return &gauge, nil
}

// btcStakingGaugeStore returns the KVStore of the gauge of total reward for
// BTC staking at each height
// prefix: BTCStakingGaugeKey
// key: gauge height
// value: gauge of rewards for BTC staking at this height
func (k Keeper) btcStakingGaugeStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.BTCStakingGaugeKey)
}
