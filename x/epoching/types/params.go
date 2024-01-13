package types

import (
	"fmt"
)

const (
	DefaultEpochInterval uint64 = 10
)

// NewParams creates a new Params instance
func NewParams(epochInterval uint64) Params {
	return Params{
		EpochInterval: epochInterval,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultEpochInterval)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEpochInterval(p.EpochInterval); err != nil {
		return err
	}

	return nil
}

func validateEpochInterval(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 2 {
		return fmt.Errorf("epoch interval must be at least 2: %d", v)
	}

	return nil
}
