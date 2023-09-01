package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress, rg *types.RewardGauge) {
	store := k.rewardGaugeStore(ctx, sType)
	rgBytes := k.cdc.MustMarshal(rg)
	store.Set(addr, rgBytes)
}

func (k Keeper) GetRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress) (*types.RewardGauge, error) {
	store := k.rewardGaugeStore(ctx, sType)
	rgBytes := store.Get(addr)
	if len(rgBytes) == 0 {
		return nil, types.ErrRewardGaugeNotFound
	}

	var rg types.RewardGauge
	k.cdc.MustUnmarshal(rgBytes, &rg)
	return &rg, nil
}

// rewardGaugeStore returns the KVStore of the reward gauge of a stakeholder
// of a given type {submitter, reporter, BTC validator, BTC delegation}
// prefix: RewardGaugeKey
// key: (stakeholder type || stakeholder address)
// value: reward gauge
func (k Keeper) rewardGaugeStore(ctx sdk.Context, sType types.StakeholderType) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	rgStore := prefix.NewStore(store, types.RewardGaugeKey)
	return prefix.NewStore(rgStore, sType.Bytes())
}
