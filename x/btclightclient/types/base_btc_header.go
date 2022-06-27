package types

import (
	bbl "github.com/babylonchain/babylon/types"
)

// NewBaseBTCHeader creates a new Params instance
func NewBaseBTCHeader(headerBytes bbl.BTCHeaderBytes, height uint64) BaseBTCHeader {
	return BaseBTCHeader{Header: &headerBytes, Height: height}
}

// DefaultBaseBTCHeader returns a default set of parameters
func DefaultBaseBTCHeader(headerBytes bbl.BTCHeaderBytes, height uint64) BaseBTCHeader {
	return NewBaseBTCHeader(headerBytes, height)
}

// Validate validates the base BTC header
func (p BaseBTCHeader) Validate() error {
	return nil
}
