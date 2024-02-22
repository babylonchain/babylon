package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVotingPowerDistCache() *VotingPowerDistCache {
	return &VotingPowerDistCache{
		TotalVotingPower:        0,
		FinalityProviders:       []*FinalityProviderDistInfo{},
		ActiveFinalityProviders: []*FinalityProviderDistInfo{},
	}
}

func (dc *VotingPowerDistCache) Empty() bool {
	return len(dc.FinalityProviders) == 0
}

func (dc *VotingPowerDistCache) AddFinalityProviderDistInfo(v *FinalityProviderDistInfo) {
	dc.FinalityProviders = append(dc.FinalityProviders, v)
}

// ApplyActiveFinalityProviders filters out the top N finality providers
// and their total voting power, and record them in the cache
func (dc *VotingPowerDistCache) ApplyActiveFinalityProviders(n uint32) {
	// reset total voting power
	dc.TotalVotingPower = 0
	// filter top N finality providers
	dc.ActiveFinalityProviders = FilterTopNFinalityProviders(dc.FinalityProviders, n)
	// construct voting power
	for _, fp := range dc.ActiveFinalityProviders {
		dc.TotalVotingPower += fp.TotalVotingPower
	}
}

// FilterVotedFinalityProviders filters out finality providers that have voted according to a map of given voters
// and update total voted power accordingly
func (dc *VotingPowerDistCache) FilterVotedFinalityProviders(voterBTCPKs map[string]struct{}) {
	filteredFps := []*FinalityProviderDistInfo{}
	totalVotingPower := uint64(0)
	for _, v := range dc.ActiveFinalityProviders {
		if _, ok := voterBTCPKs[v.BtcPk.MarshalHex()]; ok {
			filteredFps = append(filteredFps, v)
			totalVotingPower += v.TotalVotingPower
		}
	}
	dc.ActiveFinalityProviders = filteredFps
	dc.TotalVotingPower = totalVotingPower
}

// GetFinalityProviderPortion returns the portion of a finality provider's voting power out of the total voting power
func (dc *VotingPowerDistCache) GetFinalityProviderPortion(v *FinalityProviderDistInfo) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(v.TotalVotingPower)).QuoTruncate(sdkmath.LegacyNewDec(int64(dc.TotalVotingPower)))
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

func (v *FinalityProviderDistInfo) AddBTCDel(btcDel *BTCDelegation) {
	btcDelDistInfo := &BTCDelDistInfo{
		BtcPk:         btcDel.BtcPk,
		BabylonPk:     btcDel.BabylonPk,
		StakingTxHash: btcDel.MustGetStakingTxHash().String(),
		VotingPower:   btcDel.TotalSat,
	}
	v.BtcDels = append(v.BtcDels, btcDelDistInfo)
	v.TotalVotingPower += btcDelDistInfo.VotingPower
}

func (v *FinalityProviderDistInfo) AddBTCDelDistInfo(d *BTCDelDistInfo) {
	v.BtcDels = append(v.BtcDels, d)
	v.TotalVotingPower += d.VotingPower
}

// GetBTCDelPortion returns the portion of a BTC delegation's voting power out of
// the finality provider's total voting power
func (v *FinalityProviderDistInfo) GetBTCDelPortion(d *BTCDelDistInfo) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(d.VotingPower)).QuoTruncate(sdkmath.LegacyNewDec(int64(v.TotalVotingPower)))
}

func (d *BTCDelDistInfo) GetAddress() sdk.AccAddress {
	return sdk.AccAddress(d.BabylonPk.Address())
}
