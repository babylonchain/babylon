package btcutils

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// 1 byte for OP_RETURN opcode and at most 80bytes of data
	maxOpReturnPkScriptSize = 81
)

// Parsed proof represent semantically valid:
// - Bitcoin Header
// - Bitcoin Header hash
// - Bitcoin Transaction
// - Bitcoin Transaction index in block
// - Non-empty OpReturnData
type ParsedProof struct {
	BlockHeader wire.BlockHeader
	// keeping header hash to avoid recomputing it everytime
	BlockHash        chainhash.Hash
	Transaction      *btcutil.Tx
	TransactionBytes []byte
	TransactionIdx   uint32
	OpReturnData     []byte
}

// Concatenates and double hashes two provided inputs
func hashConcat(a []byte, b []byte) chainhash.Hash {
	c := []byte{}
	c = append(c, a...)
	c = append(c, b...)
	return chainhash.HashH(c)
}

// Prove checks the validity of a merkle proof
// proof logic copied from:
// https://github.com/summa-tx/bitcoin-spv/blob/fb2a61e7a941d421ae833789d97ed10d2ad79cfe/golang/btcspv/bitcoin_spv.go#L498
// main reason for not bringing library in, is that we already use btcd
// bitcoin primitives and this library defines their own which could lead
// to some mixups
func prove(tx *btcutil.Tx, merkleRoot *chainhash.Hash, intermediateNodes []byte, index uint32) bool {
	txHash := tx.Hash()

	// Shortcut the empty-block case
	if txHash.IsEqual(merkleRoot) && index == 0 && len(intermediateNodes) == 0 {
		return true
	}

	proof := []byte{}
	proof = append(proof, txHash[:]...)
	proof = append(proof, intermediateNodes...)
	proof = append(proof, merkleRoot[:]...)

	var current chainhash.Hash
	idx := index
	proofLength := len(proof)

	if proofLength%32 != 0 {
		return false
	}

	if proofLength == 64 {
		return false
	}

	root := proof[proofLength-32:]

	cur := proof[:32:32]
	copy(current[:], cur)

	numSteps := (proofLength / 32) - 1

	for i := 1; i < numSteps; i++ {
		start := i * 32
		end := i*32 + 32
		next := proof[start:end:end]
		if idx%2 == 1 {
			current = hashConcat(next, current[:])
		} else {
			current = hashConcat(current[:], next)
		}
		idx >>= 1
	}

	return bytes.Equal(current[:], root)
}

func extractOpReturnData(tx *btcutil.Tx) []byte {
	msgTx := tx.MsgTx()
	opReturnData := []byte{}

	for _, output := range msgTx.TxOut {
		pkScript := output.PkScript
		pkScriptLen := len(pkScript)
		if pkScriptLen > 0 &&
			pkScriptLen <= maxOpReturnPkScriptSize &&
			pkScript[0] == txscript.OP_RETURN {
			// drop first op return byte
			opReturnData = append(opReturnData, pkScript[1:]...)
		}
	}

	return opReturnData
}

// TODO define domain errors with nice error messages
// TODO add some tests for the proof validation
func ParseProof(
	btcTransaction []byte,
	transactionIndex uint32,
	merkleProof []byte,
	btcHeader []byte,
	powLimit *big.Int) (*ParsedProof, error) {
	tx, e := btcutil.NewTxFromBytes(btcTransaction)

	if e != nil {
		return nil, e
	}

	e = blockchain.CheckTransactionSanity(tx)

	if e != nil {
		return nil, e
	}

	header := types.BTCHeaderBytes(btcHeader).ToBlockHeader()

	e = types.ValidateHeader(header, powLimit)

	if e != nil {
		return nil, e
	}

	validProof := prove(tx, &header.MerkleRoot, merkleProof, transactionIndex)

	if !validProof {
		return nil, fmt.Errorf("header failed validation")
	}

	opReturnData := extractOpReturnData(tx)

	if len(opReturnData) == 0 {
		return nil, fmt.Errorf("provided transaction should provide op return data")
	}

	parsedProof := &ParsedProof{
		BlockHeader:      *header,
		BlockHash:        header.BlockHash(),
		Transaction:      tx,
		TransactionBytes: btcTransaction,
		TransactionIdx:   transactionIndex,
		OpReturnData:     opReturnData,
	}

	return parsedProof, nil
}
