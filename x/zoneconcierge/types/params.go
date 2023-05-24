package types

import (
	"fmt"
)

const (
	DefaultIBCPacketTimeoutMinutes uint32 = 30
)

// NewParams creates a new Params instance
func NewParams(ibcPacketTimeoutMinutes uint32) Params {
	return Params{
		IbcPacketTimeoutMinutes: ibcPacketTimeoutMinutes,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultIBCPacketTimeoutMinutes)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.IbcPacketTimeoutMinutes == 0 {
		return fmt.Errorf("IbcPacketTimeoutMinutes must be positive")
	}

	return nil
}
