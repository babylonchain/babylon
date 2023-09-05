package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) withdrawReward(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress) (sdk.Coins, error) {
	// retrieve reward gauge of the given stakeholder
	rg, err := k.GetRewardGauge(ctx, sType, addr)
	if err != nil {
		return nil, err
	}
	// get withdrawable coins
	withdrawableCoins := rg.GetWithdrawableCoins()
	if len(withdrawableCoins) == 0 {
		return nil, types.ErrNoWithdrawableCoins
	}
	// transfer withdrawable coins from incentive module account to the stakeholder's address
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, withdrawableCoins); err != nil {
		return nil, err
	}
	// empty reward gauge
	rg.Clear()
	k.SetRewardGauge(ctx, sType, addr, rg)
	// all good, return
	return withdrawableCoins, nil
}

func (k Keeper) SetRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress, rg *types.RewardGauge) {
	store := k.rewardGaugeStore(ctx, sType)
	rgBytes := k.cdc.MustMarshal(rg)
	store.Set(addr, rgBytes)
}

func (k Keeper) HasRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress) bool {
	store := k.rewardGaugeStore(ctx, sType)
	return store.Has(addr)
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
