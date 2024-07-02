package types

import (
	"fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Default parameter namespace
const (
	DefaultSignedBlocksWindow = int64(100)
	DefaultMinPubRand         = 100
)

var (
	DefaultMinSignedPerWindow = math.LegacyNewDecWithPrec(5, 1)
)

var _ paramtypes.ParamSet = (*Params)(nil)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		SignedBlocksWindow: DefaultSignedBlocksWindow,
		MinSignedPerWindow: DefaultMinSignedPerWindow,
		MinPubRand:         DefaultMinPubRand,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

func validateMinPubRand(minPubRand uint64) error {
	if minPubRand == 0 {
		return fmt.Errorf("min Pub Rand cannot be 0")
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the params
func (p Params) Validate() error {
	if err := validateSignedBlocksWindow(p.SignedBlocksWindow); err != nil {
		return err
	}
	if err := validateMinSignedPerWindow(p.MinSignedPerWindow); err != nil {
		return err
	}
	if err := validateMinPubRand(p.MinPubRand); err != nil {
		return err
	}

	return nil
}

func validateSignedBlocksWindow(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("signed blocks window must be positive: %d", v)
	}

	return nil
}

func validateMinSignedPerWindow(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("min signed per window cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("min signed per window cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("min signed per window too large: %s", v)
	}

	return nil
}

// MinSignedPerWindowInt returns min signed per window as an integer (vs the decimal in the param)
func (p *Params) MinSignedPerWindowInt() int64 {
	signedBlocksWindow := p.SignedBlocksWindow
	minSignedPerWindow := p.MinSignedPerWindow

	// NOTE: RoundInt64 will never panic as minSignedPerWindow is
	//       less than 1.
	return minSignedPerWindow.MulInt64(signedBlocksWindow).RoundInt64()
}
