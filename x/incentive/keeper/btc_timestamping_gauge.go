package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) accumulateBTCTimestampingReward(ctx sdk.Context, btcTimestampingReward sdk.Coins) {
	var (
		epoch = k.epochingKeeper.GetEpoch(ctx)
		gauge *types.Gauge
		err   error
	)

	// do nothing at epoch 0
	if epoch.EpochNumber == 0 {
		return
	}

	// update BTC timestamping reward gauge
	if k.HasBTCTimestampingGauge(ctx, epoch.EpochNumber) {
		// if this epoch already has an non-empty gauge, accumulate
		gauge, err = k.GetBTCTimestampingGauge(ctx, epoch.EpochNumber)
		if err != nil {
			panic(err) // only programming error is possible
		}
		gauge.Coins = gauge.Coins.Add(btcTimestampingReward...) // accumulate coins in the gauge
	} else {
		// if this epoch does not have a gauge yet, create a new one
		gauge = types.NewGauge(btcTimestampingReward)
	}
	k.SetBTCTimestampingGauge(ctx, epoch.EpochNumber, gauge)

	// transfer the BTC timestamping reward from fee collector account to incentive module account
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, btcTimestampingReward)
	if err != nil {
		// this can only be programming error and is unrecoverable
		panic(err)
	}
}

func (k Keeper) SetBTCTimestampingGauge(ctx sdk.Context, epoch uint64, gauge *types.Gauge) {
	store := k.btcTimestampingGaugeStore(ctx)
	gaugeBytes := k.cdc.MustMarshal(gauge)
	store.Set(sdk.Uint64ToBigEndian(epoch), gaugeBytes)
}

func (k Keeper) HasBTCTimestampingGauge(ctx sdk.Context, epoch uint64) bool {
	store := k.btcTimestampingGaugeStore(ctx)
	return store.Has(sdk.Uint64ToBigEndian(epoch))
}

func (k Keeper) GetBTCTimestampingGauge(ctx sdk.Context, epoch uint64) (*types.Gauge, error) {
	store := k.btcTimestampingGaugeStore(ctx)
	gaugeBytes := store.Get(sdk.Uint64ToBigEndian(epoch))
	if len(gaugeBytes) == 0 {
		return nil, types.ErrBTCTimestampingGaugeNotFound
	}

	var gauge types.Gauge
	k.cdc.MustUnmarshal(gaugeBytes, &gauge)
	return &gauge, nil
}

// btcTimestampingGaugeStore returns the KVStore of the gauge of total reward for
// BTC timestamping at each epoch
// prefix: BTCTimestampingGaugeKey
// key: epoch number
// value: gauge of rewards for BTC timestamping at this epoch
func (k Keeper) btcTimestampingGaugeStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.BTCTimestampingGaugeKey)
}
