package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRewardDistCache() *RewardDistCache {
	return &RewardDistCache{
		TotalVotingPower:  0,
		FinalityProviders: []*FinalityProviderDistInfo{},
	}
}

func (rdc *RewardDistCache) AddFinalityProviderDistInfo(v *FinalityProviderDistInfo) {
	if v.TotalVotingPower > 0 {
		// append finality provider dist info and accumulate voting power
		rdc.FinalityProviders = append(rdc.FinalityProviders, v)
		rdc.TotalVotingPower += v.TotalVotingPower
	}
}

// FilterVotedFinalityProviders filters out finality providers that have voted according to a map of given voters
// and update total voted power accordingly
func (rdc *RewardDistCache) FilterVotedFinalityProviders(voterBTCPKs map[string]struct{}) {
	filteredFps := []*FinalityProviderDistInfo{}
	totalVotingPower := uint64(0)
	for _, v := range rdc.FinalityProviders {
		if _, ok := voterBTCPKs[v.BtcPk.MarshalHex()]; ok {
			filteredFps = append(filteredFps, v)
			totalVotingPower += v.TotalVotingPower
		}
	}
	rdc.FinalityProviders = filteredFps
	rdc.TotalVotingPower = totalVotingPower
}

// GetFinalityProviderPortion returns the portion of a finality provider's voting power out of the total voting power
func (rdc *RewardDistCache) GetFinalityProviderPortion(v *FinalityProviderDistInfo) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(v.TotalVotingPower)).QuoTruncate(sdkmath.LegacyNewDec(int64(rdc.TotalVotingPower)))
}

func NewFinalityProviderDistInfo(fp *FinalityProvider) *FinalityProviderDistInfo {
	return &FinalityProviderDistInfo{
		BtcPk:            fp.BtcPk,
		BabylonPk:        fp.BabylonPk,
		Commission:       fp.Commission,
		TotalVotingPower: 0,
		BtcDels:          []*BTCDelDistInfo{},
	}
}

func (v *FinalityProviderDistInfo) GetAddress() sdk.AccAddress {
	return sdk.AccAddress(v.BabylonPk.Address())
}

func (v *FinalityProviderDistInfo) AddBTCDel(btcDel *BTCDelegation, btcHeight uint64, wValue uint64, covenantQuorum uint32) {
	btcDelDistInfo := &BTCDelDistInfo{
		BabylonPk:   btcDel.BabylonPk,
		VotingPower: btcDel.VotingPower(btcHeight, wValue, covenantQuorum),
	}

	if btcDelDistInfo.VotingPower > 0 {
		// if this BTC delegation has voting power, append it and accumulate voting power
		v.BtcDels = append(v.BtcDels, btcDelDistInfo)
		v.TotalVotingPower += btcDelDistInfo.VotingPower
	}
}

func (v *FinalityProviderDistInfo) AddBTCDistInfo(info *BTCDelDistInfo) {
	if info.VotingPower > 0 {
		// if this BTC delegation has voting power, append it and accumulate voting power
		v.BtcDels = append(v.BtcDels, info)
		v.TotalVotingPower += info.VotingPower
	}
}

// GetBTCDelPortion returns the portion of a BTC delegation's voting power out of
// the finality provider's total voting power
func (v *FinalityProviderDistInfo) GetBTCDelPortion(d *BTCDelDistInfo) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(d.VotingPower)).QuoTruncate(sdkmath.LegacyNewDec(int64(v.TotalVotingPower)))
}

func (d *BTCDelDistInfo) GetAddress() sdk.AccAddress {
	return sdk.AccAddress(d.BabylonPk.Address())
}
