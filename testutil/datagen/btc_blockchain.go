package datagen

import (
	"math/big"
	"math/rand"

	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func GenRandomBlock(numBabylonTxs int, prevHash *chainhash.Hash) (*wire.MsgBlock, *btctxformatter.RawBtcCheckpoint) {
	var (
		randomTxs []*wire.MsgTx
		rawCkpt   *btctxformatter.RawBtcCheckpoint
	)

	if numBabylonTxs == 2 {
		randomTxs, rawCkpt = GenRandomBabylonTxPair()
	} else if numBabylonTxs == 1 {
		randomTxs, _ = GenRandomBabylonTxPair()
		randomTxs[1] = GenRandomTx()
		rawCkpt = nil
	} else {
		randomTxs = []*wire.MsgTx{GenRandomTx(), GenRandomTx()}
		rawCkpt = nil
	}
	coinbaseTx := createCoinbaseTx(rand.Int31(), &chaincfg.SimNetParams)
	msgTxs := []*wire.MsgTx{coinbaseTx}
	msgTxs = append(msgTxs, randomTxs...)

	// calculate correct Merkle root
	merkleRoot := calcMerkleRoot(msgTxs)
	// don't apply any difficulty
	difficulty, _ := new(big.Int).SetString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	workBits := blockchain.BigToCompact(difficulty)

	header := GenRandomBtcdHeader()
	header.MerkleRoot = merkleRoot
	header.Bits = workBits
	if prevHash == nil {
		header.PrevBlock = chainhash.DoubleHashH(GenRandomByteArray(10))
	} else {
		header.PrevBlock = *prevHash
	}
	// find a nonce that satisfies difficulty
	SolveBlock(header)

	block := &wire.MsgBlock{
		Header:       *header,
		Transactions: msgTxs,
	}
	return block, rawCkpt
}

func GenRandomBlockchainWithBabylonTx(n uint64, partialPercentage float32, fullPercentage float32) ([]*wire.MsgBlock, int, []*btctxformatter.RawBtcCheckpoint) {
	blocks := []*wire.MsgBlock{}
	numCkptSegs := 0
	rawCkpts := []*btctxformatter.RawBtcCheckpoint{}
	// percentage should be [0, 1]
	if partialPercentage < 0 || partialPercentage > 1 {
		return blocks, 0, rawCkpts
	}
	if fullPercentage < 0 || fullPercentage > 1 {
		return blocks, 0, rawCkpts
	}
	// n should be > 0
	if n == 0 {
		return blocks, 0, rawCkpts
	}

	// genesis block
	genesisBlock, rawCkpt := GenRandomBlock(0, nil)
	blocks = append(blocks, genesisBlock)
	rawCkpts = append(rawCkpts, rawCkpt)

	// blocks after genesis
	for i := uint64(1); i < n; i++ {
		var msgBlock *wire.MsgBlock
		prevHash := blocks[len(blocks)-1].BlockHash()
		if rand.Float32() < partialPercentage {
			msgBlock, rawCkpt = GenRandomBlock(1, &prevHash)
			numCkptSegs += 1
		} else if rand.Float32() < partialPercentage+fullPercentage {
			msgBlock, rawCkpt = GenRandomBlock(2, &prevHash)
			numCkptSegs += 2
		} else {
			msgBlock, rawCkpt = GenRandomBlock(0, &prevHash)
		}

		blocks = append(blocks, msgBlock)
		rawCkpts = append(rawCkpts, rawCkpt)
	}
	return blocks, numCkptSegs, rawCkpts
}
