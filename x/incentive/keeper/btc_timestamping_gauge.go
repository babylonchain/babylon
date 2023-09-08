package keeper

import (
	"cosmossdk.io/math"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardBTCTimestamping distributes rewards to submitters/reporters of a checkpoint at a given epoch
// according to the reward distribution cache
func (k Keeper) RewardBTCTimestamping(ctx sdk.Context, epoch uint64, rdi *btcctypes.RewardDistInfo) {
	gauge, err := k.GetBTCTimestampingGauge(ctx, epoch)
	if err != nil {
		// failing to get a reward gauge at a finalised epoch is a programming error
		panic(err)
	}

	params := k.GetParams(ctx)
	btcTimestampingPortion := params.BTCTimestampingPortion()
	// TODO: parameterise bestPortion
	bestPortion := math.LegacyNewDecWithPrec(80, 2) // 80 * 10^{-2} = 0.8

	// distribute coins to best submitter
	submitterPortion := params.SubmitterPortion.QuoTruncate(btcTimestampingPortion)
	coinsToSubmitters := gauge.GetCoinsPortion(submitterPortion)
	coinsToBestSubmitter := types.GetCoinsPortion(coinsToSubmitters, bestPortion)
	if coinsToBestSubmitter.IsAllPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rdi.Best.Submitter, coinsToBestSubmitter); err != nil {
			// incentive module account is supposed to have enough balance
			panic(err)
		}
	}
	// distribute coins to best reporter
	reporterPortion := params.ReporterPortion.QuoTruncate(btcTimestampingPortion)
	coinsToReporters := gauge.GetCoinsPortion(reporterPortion)
	coinsToBestReporter := types.GetCoinsPortion(coinsToReporters, bestPortion)
	if coinsToBestReporter.IsAllPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rdi.Best.Reporter, coinsToBestReporter); err != nil {
			// incentive module account is supposed to have enough balance
			panic(err)
		}
	}

	// if there is only 1 submission, distribute the rest to submitter and reporter, then skip the rest logic
	if len(rdi.Others) == 0 {
		// give rest coins to the best submitter
		restCoinsToSubmitter := coinsToSubmitters.Sub(coinsToBestSubmitter...)
		if restCoinsToSubmitter.IsAllPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rdi.Best.Submitter, restCoinsToSubmitter); err != nil {
				// incentive module account is supposed to have enough balance
				panic(err)
			}
		}
		// give rest coins to the best reporter
		restCoinsToReporter := coinsToReporters.Sub(coinsToBestReporter...)
		if restCoinsToReporter.IsAllPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, rdi.Best.Reporter, restCoinsToReporter); err != nil {
				// incentive module account is supposed to have enough balance
				panic(err)
			}
		}
		// skip the rest logic
		return
	}

	// distribute the rest to each of the other submitters
	// TODO: our tokenomics might specify weights for the rest submitters in the future
	coinsToOtherSubmitters := coinsToSubmitters.Sub(coinsToBestSubmitter...)
	eachOtherSubmitterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
	coinsToEachOtherSubmitter := types.GetCoinsPortion(coinsToOtherSubmitters, eachOtherSubmitterPortion)
	if coinsToEachOtherSubmitter.IsAllPositive() {
		for _, submission := range rdi.Others {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, submission.Submitter, coinsToEachOtherSubmitter); err != nil {
				// incentive module account is supposed to have enough balance
				panic(err)
			}
		}
	}

	// distribute the rest to each of the other reporters
	// TODO: our tokenomics might specify weights for the rest reporters in the future
	coinsToOtherReporters := coinsToReporters.Sub(coinsToBestReporter...)
	eachOtherReporterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
	coinsToEachOtherReporter := types.GetCoinsPortion(coinsToOtherReporters, eachOtherReporterPortion)
	if coinsToEachOtherReporter.IsAllPositive() {
		for _, submission := range rdi.Others {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, submission.Reporter, coinsToEachOtherReporter); err != nil {
				// incentive module account is supposed to have enough balance
				panic(err)
			}
		}
	}
}

func (k Keeper) accumulateBTCTimestampingReward(ctx sdk.Context, btcTimestampingReward sdk.Coins) {
	var (
		epoch = k.epochingKeeper.GetEpoch(ctx)
		gauge *types.Gauge
		err   error
	)

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
