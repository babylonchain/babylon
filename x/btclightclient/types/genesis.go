package types

import (
	"errors"
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
		BtcHeaders: []*BTCHeaderInfo{&defaultBaseHeader},
		Params:     DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params in genesis: %w", err)
	}

	// Initial btc header serves as de-facto genesis header for the module.
	// At least one BTC Header is needed to apply all validation rules to the next headers.
	// If we don't have an initial btc header, we cannot validate the rules on the next (as it has no parent).
	// All following headers that are to be inserted in chain are going to be validated based on the previous ones.
	// (all parent-child relantionships, all difficulty transitions).
	if len(gs.BtcHeaders) == 0 {
		// if you have no initial header, you can't validate the following ones.
		return errors.New("no btc header set on genesis")
	}

	// We Require that genesis block is difficulty adjustment block, so that we can
	// properly calculate the difficulty adjustments in the future.
	// TODO: Even though number of block per re-target depends on the network, in reality it
	// is always 2016. Maybe we should consider moving it to param, or try to pass
	// it through
	// isRetarget := IsRetargetBlock(gs.BtcHeaders[0], &chaincfg.SimNetParams)
	// if !isRetarget {
	// 	return fmt.Errorf("genesis block must be a difficulty adjustment block")
	// }

	// for _, header := range gs.BtcHeaders {
	// 	if err := header.Validate(); err != nil {
	// 		return err
	// 	}
	// }
	// TODO: validate headers have proper parent-child relationships and proper proof of work

	return nil
}
