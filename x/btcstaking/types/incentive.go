package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRewardDistCache() *RewardDistCache {
	return &RewardDistCache{
		TotalVotingPower: 0,
		BtcVals:          []*BTCValDistInfo{},
	}
}

func (rdc *RewardDistCache) AddBTCValDistInfo(v *BTCValDistInfo) {
	if v.TotalVotingPower > 0 {
		// append BTC validator dist info and accumulate voting power
		rdc.BtcVals = append(rdc.BtcVals, v)
		rdc.TotalVotingPower += v.TotalVotingPower
	}
}

// FilterVotedBTCVals filters out BTC validators that have voted according to a map of given voters
// and update total voted power accordingly
func (rdc *RewardDistCache) FilterVotedBTCVals(voterBTCPKs map[string]struct{}) {
	filteredBTCVals := []*BTCValDistInfo{}
	totalVotingPower := uint64(0)
	for _, v := range rdc.BtcVals {
		if _, ok := voterBTCPKs[v.BtcPk.MarshalHex()]; ok {
			filteredBTCVals = append(filteredBTCVals, v)
			totalVotingPower += v.TotalVotingPower
		}
	}
	rdc.BtcVals = filteredBTCVals
	rdc.TotalVotingPower = totalVotingPower
}

// GetBTCValPortion returns the portion of a BTC validator's voting power out of the total voting power
func (rdc *RewardDistCache) GetBTCValPortion(v *BTCValDistInfo) sdk.Dec {
	return math.LegacyNewDec(int64(v.TotalVotingPower)).QuoTruncate(math.LegacyNewDec(int64(rdc.TotalVotingPower)))
}

func NewBTCValDistInfo(btcVal *BTCValidator) *BTCValDistInfo {
	return &BTCValDistInfo{
		BtcPk:            btcVal.BtcPk,
		BabylonPk:        btcVal.BabylonPk,
		Commission:       btcVal.Commission,
		TotalVotingPower: 0,
		BtcDels:          []*BTCDelDistInfo{},
	}
}

func (v *BTCValDistInfo) GetAddress() sdk.AccAddress {
	return sdk.AccAddress(v.BabylonPk.Address())
}

func (v *BTCValDistInfo) AddBTCDel(btcDel *BTCDelegation, btcHeight uint64, wValue uint64) {
	btcDelDistInfo := &BTCDelDistInfo{
		BabylonPk:   btcDel.BabylonPk,
		VotingPower: btcDel.VotingPower(btcHeight, wValue),
	}

	if btcDelDistInfo.VotingPower > 0 {
		// if this BTC delegation has voting power, append it and accumulate voting power
		v.BtcDels = append(v.BtcDels, btcDelDistInfo)
		v.TotalVotingPower += btcDelDistInfo.VotingPower
	}
}

// GetBTCValPortion returns the portion of a BTC delegation's voting power out of
// the BTC validator's total voting power
func (v *BTCValDistInfo) GetBTCDelPortion(d *BTCDelDistInfo) sdk.Dec {
	return math.LegacyNewDec(int64(d.VotingPower)).QuoTruncate(math.LegacyNewDec(int64(v.TotalVotingPower)))
}

func (d *BTCDelDistInfo) GetAddress() sdk.AccAddress {
	return sdk.AccAddress(d.BabylonPk.Address())
}
