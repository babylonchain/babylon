package types

import (
	"fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		SubmitterPortion:  math.LegacyNewDecWithPrec(5, 2), // 5 * 10^{-2} = 0.05
		ReporterPortion:   math.LegacyNewDecWithPrec(5, 2), // 5 * 10^{-2} = 0.05
		BtcStakingPortion: math.LegacyNewDecWithPrec(2, 1), // 2 * 10^{-1} = 0.2
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// TotalPortion calculates the sum of portions of all stakeholders
func (p *Params) TotalPortion() math.LegacyDec {
	sum := p.SubmitterPortion
	sum = sum.Add(p.ReporterPortion)
	sum = sum.Add(p.BtcStakingPortion)
	return sum
}

// BTCTimestampingPortion calculates the sum of portions of all BTC timestamping stakeholders
func (p *Params) BTCTimestampingPortion() math.LegacyDec {
	sum := p.SubmitterPortion
	sum = sum.Add(p.ReporterPortion)
	return sum
}

// BTCStakingPortion calculates the sum of portions of all BTC staking stakeholders
func (p *Params) BTCStakingPortion() math.LegacyDec {
	return p.BtcStakingPortion
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.SubmitterPortion.IsNil() {
		return fmt.Errorf("SubmitterPortion should not be nil")
	}
	if p.ReporterPortion.IsNil() {
		return fmt.Errorf("ReporterPortion should not be nil")
	}
	if p.BtcStakingPortion.IsNil() {
		return fmt.Errorf("BtcStakingPortion should not be nil")
	}

	// sum of all portions should be less than 1
	if p.TotalPortion().GTE(math.LegacyOneDec()) {
		return fmt.Errorf("sum of all portions should be less than 1")
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
