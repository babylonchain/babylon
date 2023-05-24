package types

import (
	"fmt"
)

const (
	DefaultIbcPacketTimeoutSeconds uint32 = 60 * 60 * 24       // 24 hours
	MaxIbcPacketTimeoutSeconds     uint32 = 60 * 60 * 24 * 365 // 1 year
)

// NewParams creates a new Params instance
func NewParams(ibcPacketTimeoutSeconds uint32) Params {
	return Params{
		IbcPacketTimeoutSeconds: ibcPacketTimeoutSeconds,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultIbcPacketTimeoutSeconds)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.IbcPacketTimeoutSeconds == 0 {
		return fmt.Errorf("IbcPacketTimeoutSeconds must be positive")
	}
	if p.IbcPacketTimeoutSeconds > MaxIbcPacketTimeoutSeconds {
		return fmt.Errorf("IbcPacketTimeoutSeconds must be no larger than %d", MaxIbcPacketTimeoutSeconds)
	}

	return nil
}
