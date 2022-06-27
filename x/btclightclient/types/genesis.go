package types

import (
	bbl "github.com/babylonchain/babylon/types"
)

// TODO: get these from a configuration file
const (
	DefaultBaseHeaderHex    string = "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22"
	DefaultBaseHeaderHeight uint64 = 736056
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	headerBytes, _ := bbl.NewBTCHeaderBytesFromHex(DefaultBaseHeaderHex)

	return &GenesisState{
		Params:        DefaultParams(),
		BaseBtcHeader: DefaultBaseBTCHeader(headerBytes, DefaultBaseHeaderHeight),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	err = gs.BaseBtcHeader.Validate()
	if err != nil {
		return err
	}
	return nil
}
