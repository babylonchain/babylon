package types

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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

func GetBaseBTCHeaderHex() string {
	// TODO: get this from a configuration file
	hex := "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22"
	return hex
}

func GetBaseBTCHeaderHeight() uint64 {
	// TODO: get this from a configuration file
	height := uint64(736056)
	return height
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

func GetBaseBTCHeaderBytes() BTCHeaderBytes {
	hex := GetBaseBTCHeaderHex()
	headerBytes, err := NewBTCHeaderBytesFromHex(hex)
	if err != nil {
		panic("Base BTC header hex cannot be converted to bytes")
	}
	return headerBytes
}

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func createBranch(nodes []*chainhash.Hash, numfLeafs uint, idx uint) []*chainhash.Hash {

	var branch []*chainhash.Hash

	var size = numfLeafs
	var index = idx

	var i uint = 0

	for size > 1 {
		j := min(index^1, size-1)

		branch = append(branch, nodes[i+j])

		index = index >> 1

		i = i + size

		size = (size + 1) >> 1
	}

	return branch
}

// quite inefficiet method of calculating merkle proofs, created for testing purposes
func CreateProofForIdx(transactions [][]byte, idx uint) ([]*chainhash.Hash, error) {
	if len(transactions) == 0 {
		return nil, errors.New("can't calculate proof for empty transaction list")
	}

	if int(idx) >= len(transactions) {
		return nil, errors.New("provided index should be smaller that lenght of transaction list")
	}

	var txs []*btcutil.Tx
	for _, b := range transactions {
		tx, e := btcutil.NewTxFromBytes(b)

		if e != nil {
			return nil, e
		}

		txs = append(txs, tx)
	}

	store := blockchain.BuildMerkleTreeStore(txs, false)

	var storeNoNil []*chainhash.Hash

	// to correctly calculate indexes we need to filter out all nil hashes which
	// represents nodes which are empty
	for _, h := range store {
		if h != nil {
			storeNoNil = append(storeNoNil, h)
		}
	}

	branch := createBranch(storeNoNil, uint(len(transactions)), idx)

	return branch, nil
}
