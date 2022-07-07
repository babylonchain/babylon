package types

import (
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
	"math/big"
)

func CalcWork(header *wire.BlockHeader) *big.Int {
	return blockchain.CalcWork(header.Bits)
}

func CumulativeWork(childWork *big.Int, parentWork *big.Int) *big.Int {
	sum := new(big.Int)
	sum.Add(childWork, parentWork)
	return sum
}
