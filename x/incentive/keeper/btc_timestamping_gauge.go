package keeper

import (
	"context"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardBTCTimestamping distributes rewards to submitters/reporters of a checkpoint at a given epoch
// according to the reward distribution cache
func (k Keeper) RewardBTCTimestamping(ctx context.Context, epoch uint64, rdi *btcctypes.RewardDistInfo) {
	gauge := k.GetBTCTimestampingGauge(ctx, epoch)
	if gauge == nil {
		// failing to get a reward gauge at a finalised epoch is a programming error
		panic("failed to get a reward gauge at a finalized epoch")
	}

	params := k.GetParams(ctx)
	btcTimestampingPortion := params.BTCTimestampingPortion()
	// TODO: parameterise bestPortion
	bestPortion := math.LegacyNewDecWithPrec(80, 2) // 80 * 10^{-2} = 0.8

	// distribute coins to best submitter
	submitterPortion := params.SubmitterPortion.QuoTruncate(btcTimestampingPortion)
	coinsToSubmitters := gauge.GetCoinsPortion(submitterPortion)
	coinsToBestSubmitter := types.GetCoinsPortion(coinsToSubmitters, bestPortion)
	k.accumulateRewardGauge(ctx, types.SubmitterType, rdi.Best.Submitter, coinsToBestSubmitter)
	restCoinsToSubmitters := coinsToSubmitters.Sub(coinsToBestSubmitter...)

	// distribute coins to best reporter
	reporterPortion := params.ReporterPortion.QuoTruncate(btcTimestampingPortion)
	coinsToReporters := gauge.GetCoinsPortion(reporterPortion)
	coinsToBestReporter := types.GetCoinsPortion(coinsToReporters, bestPortion)
	k.accumulateRewardGauge(ctx, types.ReporterType, rdi.Best.Reporter, coinsToBestReporter)
	restCoinsToReporters := coinsToReporters.Sub(coinsToBestReporter...)

	// if there is only 1 submission, distribute the rest to submitter and reporter, then skip the rest logic
	if len(rdi.Others) == 0 {
		// give rest coins to the best submitter
		k.accumulateRewardGauge(ctx, types.SubmitterType, rdi.Best.Submitter, restCoinsToSubmitters)
		// give rest coins to the best reporter
		k.accumulateRewardGauge(ctx, types.ReporterType, rdi.Best.Reporter, restCoinsToReporters)
		// skip the rest logic
		return
	}

	// distribute the rest to each of the other submitters
	// TODO: our tokenomics might specify weights for the rest submitters in the future
	eachOtherSubmitterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
	coinsToEachOtherSubmitter := types.GetCoinsPortion(restCoinsToSubmitters, eachOtherSubmitterPortion)
	if coinsToEachOtherSubmitter.IsAllPositive() {
		for _, submission := range rdi.Others {
			k.accumulateRewardGauge(ctx, types.SubmitterType, submission.Submitter, coinsToEachOtherSubmitter)
		}
	}

	// distribute the rest to each of the other reporters
	// TODO: our tokenomics might specify weights for the rest reporters in the future
	eachOtherReporterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
	coinsToEachOtherReporter := types.GetCoinsPortion(restCoinsToReporters, eachOtherReporterPortion)
	if coinsToEachOtherReporter.IsAllPositive() {
		for _, submission := range rdi.Others {
			k.accumulateRewardGauge(ctx, types.ReporterType, submission.Reporter, coinsToEachOtherReporter)
		}
	}
}

func (k Keeper) accumulateBTCTimestampingReward(ctx context.Context, btcTimestampingReward sdk.Coins) {
	epoch := k.epochingKeeper.GetEpoch(ctx)

	// update BTC timestamping reward gauge
	gauge := k.GetBTCTimestampingGauge(ctx, epoch.EpochNumber)
	if gauge == nil {
		// if this epoch does not have a gauge yet, create a new one
		gauge = types.NewGauge(btcTimestampingReward...)
	} else {
		// if this epoch already has a gauge, accumulate coins in the gauge
		gauge.Coins = gauge.Coins.Add(btcTimestampingReward...)
	}

	k.SetBTCTimestampingGauge(ctx, epoch.EpochNumber, gauge)

	// transfer the BTC timestamping reward from fee collector account to incentive module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, btcTimestampingReward)
	if err != nil {
		// this can only be programming error and is unrecoverable
		panic(err)
	}
}

func (k Keeper) SetBTCTimestampingGauge(ctx context.Context, epoch uint64, gauge *types.Gauge) {
	store := k.btcTimestampingGaugeStore(ctx)
	gaugeBytes := k.cdc.MustMarshal(gauge)
	store.Set(sdk.Uint64ToBigEndian(epoch), gaugeBytes)
}

func (k Keeper) GetBTCTimestampingGauge(ctx context.Context, epoch uint64) *types.Gauge {
	store := k.btcTimestampingGaugeStore(ctx)
	gaugeBytes := store.Get(sdk.Uint64ToBigEndian(epoch))
	if gaugeBytes == nil {
		return nil
	}

	var gauge types.Gauge
	k.cdc.MustUnmarshal(gaugeBytes, &gauge)
	return &gauge
}

// btcTimestampingGaugeStore returns the KVStore of the gauge of total reward for
// BTC timestamping at each epoch
// prefix: BTCTimestampingGaugeKey
// key: epoch number
// value: gauge of rewards for BTC timestamping at this epoch
func (k Keeper) btcTimestampingGaugeStore(ctx context.Context) prefix.Store {
	storeAdaptor := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdaptor, types.BTCTimestampingGaugeKey)
}
