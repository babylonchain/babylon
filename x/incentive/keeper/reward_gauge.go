package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) withdrawReward(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress) (sdk.Coins, error) {
	// retrieve reward gauge of the given stakeholder
	rg := k.GetRewardGauge(ctx, sType, addr)
	if rg == nil {
		return nil, types.ErrRewardGaugeNotFound
	}
	// get withdrawable coins
	withdrawableCoins := rg.GetWithdrawableCoins()
	if !withdrawableCoins.IsAllPositive() {
		return nil, types.ErrNoWithdrawableCoins
	}
	// transfer withdrawable coins from incentive module account to the stakeholder's address
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, withdrawableCoins); err != nil {
		return nil, err
	}
	// empty reward gauge
	rg.SetFullyWithdrawn()
	k.SetRewardGauge(ctx, sType, addr, rg)
	// all good, return
	return withdrawableCoins, nil
}

// accumulateRewardGauge accumulates the given reward of of a given stakeholder in a given type
func (k Keeper) accumulateRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress, reward sdk.Coins) {
	// if reward contains nothing, do nothing
	if !reward.IsAllPositive() {
		return
	}
	// get reward gauge, or create a new one if it does not exist
	rg := k.GetRewardGauge(ctx, sType, addr)
	if rg == nil {
		rg = types.NewRewardGauge()
	}
	// add the given reward to reward gauge
	rg.Add(reward)
	// set back
	k.SetRewardGauge(ctx, sType, addr, rg)
}

func (k Keeper) SetRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress, rg *types.RewardGauge) {
	store := k.rewardGaugeStore(ctx, sType)
	rgBytes := k.cdc.MustMarshal(rg)
	store.Set(addr.Bytes(), rgBytes)
}

func (k Keeper) GetRewardGauge(ctx sdk.Context, sType types.StakeholderType, addr sdk.AccAddress) *types.RewardGauge {
	store := k.rewardGaugeStore(ctx, sType)
	rgBytes := store.Get(addr.Bytes())
	if rgBytes == nil {
		return nil
	}

	var rg types.RewardGauge
	k.cdc.MustUnmarshal(rgBytes, &rg)
	return &rg
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
