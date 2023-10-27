package types

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

// ValidateBTCHeader
// Perform the checks that [checkBlockHeaderSanity](https://github.com/btcsuite/btcd/blob/master/blockchain/validate.go#L430) of btcd does
//
// We skip the "timestamp should not be 2 hours into the future" check
// since this might introduce undeterministic behavior
func ValidateBTCHeader(header *wire.BlockHeader, powLimit *big.Int) error {
	msgBlock := &wire.MsgBlock{Header: *header}
	block := btcutil.NewBlock(msgBlock)

	// The upper limit for the power to be spent
	// Use the one maintained by btcd
	err := blockchain.CheckProofOfWork(block, powLimit)

	if err != nil {
		return err
	}

	if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
		str := fmt.Sprintf("block timestamp of %v has a higher "+
			"precision than one second", header.Timestamp)
		return errors.New(str)
	}

	return nil
}

func GetMaxDifficulty() big.Int {
	// Maximum btc difficulty possible
	// Use it to set the difficulty bits of blocks as well as the upper PoW limit
	// since the block hash needs to be below that
	// This is the maximum allowed given the 2^23-1 precision
	maxDifficulty := new(big.Int)
	maxDifficulty, success := maxDifficulty.SetString("ffff000000000000000000000000000000000000000000000000000000000000", 16)
	if !success {
		panic("Conversion did not succeed")
	}
	return *maxDifficulty
}
