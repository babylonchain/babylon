package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVotingPowerDistCache() *VotingPowerDistCache {
	return &VotingPowerDistCache{
		TotalVotingPower:  0,
		FinalityProviders: []*FinalityProviderDistInfo{},
	}
}

func (dc *VotingPowerDistCache) Empty() bool {
	return len(dc.FinalityProviders) == 0
}

// ProcessAllPowerDistUpdateEvents processes all events that affect
// voting power distribution, including
// - newly active BTC delegations
// - expired BTC delegations
// - slashed finality providers
func (dc *VotingPowerDistCache) ProcessAllPowerDistUpdateEvents(events []*EventPowerDistUpdate) {
	// a map where key is finality provider's BTC PK hex and value is a list of events
	// of BTC delegations that newly have voting power under this provider
	newBTCDels := map[string][]*EventBTCDelegationStateUpdate{}
	// a map where key is expired BTC delegation's staking tx hash
	expiredBTCDels := map[string]struct{}{}
	// a map where key is slashed finality providers' BTC PK
	slashedFPs := map[string]struct{}{}

	// filter and classify all events that we care about
	for _, event := range events {
		switch typedEvent := event.Ev.(type) {
		case *EventPowerDistUpdate_BtcDelStateUpdate:
			delEvent := typedEvent.BtcDelStateUpdate
			if delEvent.NewState == BTCDelegationStatus_ACTIVE {
				for _, fpBTCPK := range delEvent.FpBtcPkList {
					fpBTCPKHex := fpBTCPK.MarshalHex()
					newBTCDels[fpBTCPKHex] = append(newBTCDels[fpBTCPKHex], delEvent)
				}
			} else if delEvent.NewState == BTCDelegationStatus_UNBONDED {
				expiredBTCDels[delEvent.StakingTxHash] = struct{}{}
			}
		case *EventPowerDistUpdate_SlashedFp:
			slashedFPs[typedEvent.SlashedFp.Pk.MarshalHex()] = struct{}{}
		}
	}

	// if no voting power update, return directly
	noUpdate := len(newBTCDels) == 0 && len(expiredBTCDels) == 0 && len(slashedFPs) == 0
	if noUpdate {
		return
	}

	// at this point, there is voting power update, construct a voting power dist cache
	newDc := NewVotingPowerDistCache()

	// iterate over all finality providers to compute BTC delegations
	dc.TotalVotingPower = 0
	for i := range dc.FinalityProviders {
		// initialise the finality provider
		fp := *dc.FinalityProviders[i]
		fp.TotalVotingPower = 0
		fp.BtcDels = []*BTCDelDistInfo{}

		// if this finality provider is slashed, continue to avoid recording it
		fpBTCPKHex := fp.BtcPk.MarshalHex()
		if _, ok := slashedFPs[fpBTCPKHex]; ok {
			continue
		}

		// add all non-expired BTC delegations to the new fp
		for j := range dc.FinalityProviders[i].BtcDels {
			btcDel := *dc.FinalityProviders[i].BtcDels[j]
			if _, ok := expiredBTCDels[btcDel.StakingTxHash]; !ok {
				fp.addBTCDelDistInfo(&btcDel)
			}
		}

		// process all new BTC delegations under this finality provider
		if newBTCDelEvents, ok := newBTCDels[fpBTCPKHex]; ok {
			// handle new BTC delegations for this finality provider
			for _, e := range newBTCDelEvents {
				fp.addBTCDelDistInfo(&BTCDelDistInfo{
					BtcPk:       e.BtcPk,
					BabylonPk:   e.BabylonPk,
					VotingPower: e.TotalSat,
				})
			}
			// remove the finality provider entry in newBTCDels map, so that
			// after the for loop the rest entries in newBTCDels belongs to new
			// finality providers with new BTC delegations
			delete(newBTCDels, fpBTCPKHex)
		}

		// add this finality provider to the new cache
		newDc.AddFinalityProviderDistInfo(&fp)
	}

	// TODO: process new BTC delegations under new finality providers in newBTCDels
}

func (dc *VotingPowerDistCache) AddFinalityProviderDistInfo(v *FinalityProviderDistInfo) {
	if v.TotalVotingPower > 0 {
		// append finality provider dist info and accumulate voting power
		dc.FinalityProviders = append(dc.FinalityProviders, v)
		dc.TotalVotingPower += v.TotalVotingPower
	}
}

// FilterVotedFinalityProviders filters out finality providers that have voted according to a map of given voters
// and update total voted power accordingly
func (dc *VotingPowerDistCache) FilterVotedFinalityProviders(voterBTCPKs map[string]struct{}) {
	filteredFps := []*FinalityProviderDistInfo{}
	totalVotingPower := uint64(0)
	for _, v := range dc.FinalityProviders {
		if _, ok := voterBTCPKs[v.BtcPk.MarshalHex()]; ok {
			filteredFps = append(filteredFps, v)
			totalVotingPower += v.TotalVotingPower
		}
	}
	dc.FinalityProviders = filteredFps
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

func (v *FinalityProviderDistInfo) AddBTCDel(btcDel *BTCDelegation, btcHeight uint64, wValue uint64, covenantQuorum uint32) {
	btcDelDistInfo := &BTCDelDistInfo{
		BtcPk:         btcDel.BtcPk,
		BabylonPk:     btcDel.BabylonPk,
		StakingTxHash: btcDel.MustGetStakingTxHash().String(),
		VotingPower:   btcDel.VotingPower(btcHeight, wValue, covenantQuorum),
	}

	if btcDelDistInfo.VotingPower > 0 {
		// if this BTC delegation has voting power, append it and accumulate voting power
		v.BtcDels = append(v.BtcDels, btcDelDistInfo)
		v.TotalVotingPower += btcDelDistInfo.VotingPower
	}
}

func (v *FinalityProviderDistInfo) addBTCDelDistInfo(d *BTCDelDistInfo) {
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
