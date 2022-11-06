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

// GenRandomBtcdBlock generates a random BTC block, which can include Babylon txs.
// Specifically,
// - when numBabylonTxs == 0 or numBabylonTxs > 2, it generates a BTC block with 3 random txs.
// - when numBabylonTxs == 1, it generates a BTC block with 2 random txs and a Babylon tx.
// - when numBabylonTxs == 2, it generates a BTC block with 1 random tx and 2 Babylon txs that constitute a raw BTC checkpoint.
// When numBabylonTxs == 2, the function will return the BTC raw checkpoint as well.
func GenRandomBtcdBlock(numBabylonTxs int, prevHash *chainhash.Hash) (*wire.MsgBlock, *btctxformatter.RawBtcCheckpoint) {
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
	if prevHash != nil {
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

// GenRandomBtcdBlockchainWithBabylonTx generates a blockchain of `n` blocks, in which each block has some probability of including some Babylon txs
// Specifically, each block
// - has `oneTxThreshold` probability of including 1 Babylon tx that does not has any match
// - has `twoTxThreshold - oneTxThreshold` probability of including 2 Babylon txs that constitute a checkpoint
// - has `1 - twoTxThreshold` probability of including no Babylon tx
func GenRandomBtcdBlockchainWithBabylonTx(n uint64, oneTxThreshold float32, twoTxThreshold float32) ([]*wire.MsgBlock, int, []*btctxformatter.RawBtcCheckpoint) {
	blocks := []*wire.MsgBlock{}
	numCkptSegs := 0
	rawCkpts := []*btctxformatter.RawBtcCheckpoint{}
	if oneTxThreshold < 0 || oneTxThreshold > 1 {
		panic("oneTxThreshold should be [0, 1]")
	}
	if twoTxThreshold < oneTxThreshold || twoTxThreshold > 1 {
		panic("fullPercentage should be [oneTxThreshold, 1]")
	}
	if n == 0 {
		panic("n should be > 0")
	}

	// genesis block
	genesisBlock, rawCkpt := GenRandomBtcdBlock(0, nil)
	blocks = append(blocks, genesisBlock)
	rawCkpts = append(rawCkpts, rawCkpt)

	// blocks after genesis
	for i := uint64(1); i < n; i++ {
		var msgBlock *wire.MsgBlock
		prevHash := blocks[len(blocks)-1].BlockHash()
		if rand.Float32() < oneTxThreshold {
			msgBlock, rawCkpt = GenRandomBtcdBlock(1, &prevHash)
			numCkptSegs += 1
		} else if rand.Float32() < twoTxThreshold {
			msgBlock, rawCkpt = GenRandomBtcdBlock(2, &prevHash)
			numCkptSegs += 2
		} else {
			msgBlock, rawCkpt = GenRandomBtcdBlock(0, &prevHash)
		}

		blocks = append(blocks, msgBlock)
		rawCkpts = append(rawCkpts, rawCkpt)
	}
	return blocks, numCkptSegs, rawCkpts
}

// GenRandomBtcdHash generates a random hash in type `chainhash.Hash`, without any hash operations
func GenRandomBtcdHash() chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(GenRandomHexStr(32))
	if err != nil {
		panic(err)
	}
	return *hash
}
