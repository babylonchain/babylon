package types

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg"
)

func SimnetGenesisBlock() BTCHeaderInfo {
	// By default we use the genesis block of the simnet, as it is the best for testing
	var header = chaincfg.SimNetParams.GenesisBlock.Header
	var headerHash = chaincfg.SimNetParams.GenesisHash

	bytes := bbn.NewBTCHeaderBytesFromBlockHeader(&header)
	hash := bbn.NewBTCHeaderHashBytesFromChainhash(headerHash)
	work := CalcWork(&bytes)

	return *NewBTCHeaderInfo(
		&bytes,
		&hash,
		0,
		&work,
	)
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	defaultBaseHeader := SimnetGenesisBlock()

	return &GenesisState{
		BaseBtcHeader: defaultBaseHeader,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// We Require that genesis block is difficulty adjustment block, so that we can
	// properly calculate the difficulty adjustments in the future.
	// TODO: Even though number of block per re-target depends on the network, in reality it
	// is always 2016. Maybe we should consider moving it to param, or try to pass
	// it through
	isRetarget := IsRetargetBlock(&gs.BaseBtcHeader, &chaincfg.MainNetParams)

	if !isRetarget {
		return fmt.Errorf("genesis block must be a difficulty adjustment block")
	}

	return nil
}
