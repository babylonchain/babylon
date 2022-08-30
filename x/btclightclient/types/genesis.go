package types

import (
	bbn "github.com/babylonchain/babylon/types"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	headerBytes := bbn.GetBaseBTCHeaderBytes()
	headerHeight := bbn.GetBaseBTCHeaderHeight()
	headerHash := headerBytes.Hash()
	// The cumulative work for the Base BTC header is only the work
	// for that particular header. This means that it is very important
	// that no forks will happen that discard the base header because we
	// will not be able to detect those. Cumulative work will build based
	// on the sum of the work of the chain starting from the base header.
	headerWork := CalcWork(&headerBytes)
	baseHeaderInfo := NewBTCHeaderInfo(&headerBytes, headerHash, headerHeight, &headerWork)

	return &GenesisState{
		Params:        DefaultParams(),
		BaseBtcHeader: *baseHeaderInfo,
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
