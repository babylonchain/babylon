package types

import (
	fmt "fmt"
)

const (
	DefaultBtcConfirmationDepth          uint64 = 10
	DefaultCheckpointFinalizationTimeout uint64 = 100
)

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

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateBtcConfirmationDepth(p.BtcConfirmationDepth); err != nil {
		return err
	}
	if err := validateCheckpointFinalizationTimeout(p.CheckpointFinalizationTimeout); err != nil {
		return err
	}
	if p.BtcConfirmationDepth >= p.CheckpointFinalizationTimeout {
		return fmt.Errorf("BtcConfirmationDepth should be smaller than CheckpointFinalizationTimeout")
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
		return fmt.Errorf("CheckpointFinalizationTimeout must be positive: %d", v)
	}

	return nil
}
