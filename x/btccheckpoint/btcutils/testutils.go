package btcutils

import (
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BtcHasher struct{}

func NewBtcHasher() *BtcHasher {
	return &BtcHasher{}
}

func (h *BtcHasher) Hash(data []byte) []byte {
	return chainhash.DoubleHashB(data)
}

func HashFromString(s string) *chainhash.Hash {
	hash, e := chainhash.NewHashFromStr(s)
	if e != nil {
		panic("Invalid hex sting")
	}

	return hash
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
func createProofForIdx(transactions [][]byte, idx uint) []*chainhash.Hash {
	var txs []*btcutil.Tx
	for _, b := range transactions {
		tx, _ := parseTransaction(b)
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
	return branch
}
