package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	DefaultBtcConfirmationDepth          uint64 = 10
	DefaultCheckpointFinalizationTimeout uint64 = 100
)

var (
	KeyBtcConfirmationDepth          = []byte("BtcConfirmationDepth")
	KeyCheckpointFinalizationTimeout = []byte("CheckpointFinalizationTimeout")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(btcConfirmationDepth uint64, checkpointFinalizationTimeout uint64) Params {
	return Params{
		BtcConfirmationDepth:          btcConfirmationDepth,
		CheckpointFinalizationTimeout: checkpointFinalizationTimeout,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultBtcConfirmationDepth,
		DefaultCheckpointFinalizationTimeout,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBtcConfirmationDepth, &p.BtcConfirmationDepth, validateBtcConfirmationDepth),
		paramtypes.NewParamSetPair(KeyCheckpointFinalizationTimeout, &p.CheckpointFinalizationTimeout, validateCheckpointFinalizationTimeout),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateBtcConfirmationDepth(p.BtcConfirmationDepth); err != nil {
		return err
	}
	if err := validateCheckpointFinalizationTimeout(p.CheckpointFinalizationTimeout); err != nil {
		return err
	}

	return nil
}

func validateBtcConfirmationDepth(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("BtcConfirmationDepth must be positive: %d", v)
	}

	return nil
}

func validateCheckpointFinalizationTimeout(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("BtcConfirmationDepth must be positive: %d", v)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
