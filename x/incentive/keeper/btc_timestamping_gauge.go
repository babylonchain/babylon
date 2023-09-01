package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetBTCTimestampingGauge(ctx sdk.Context, epoch uint64, gauge *types.Gauge) {
	store := k.btcTimestampingGaugeStore(ctx)
	gaugeBytes := k.cdc.MustMarshal(gauge)
	store.Set(sdk.Uint64ToBigEndian(epoch), gaugeBytes)
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
