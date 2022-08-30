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
)

const (
	// 1 byte for OP_RETURN opcode
	// 1 byte for OP_DATAXX, or 2 bytes for OP_PUSHDATA1 opcode
	// max 80 bytes of application specific data
	// This stems from the fact that if data in op_return is less than 75 bytes
	// one of OP_DATAXX opcodes is used (https://wiki.bitcoinsv.io/index.php/Pushdata_Opcodes#Opcodes_1-75_.280x01_-_0x4B.29)
	// but if data in op_return is between 76 and 80bytes, OP_PUSHDATA1 needs to be used
	// in which 1 byte indicates op code itself and 1 byte indicates how many bytes
	// are pushed onto stack (https://wiki.bitcoinsv.io/index.php/Pushdata_Opcodes#OP_PUSHDATA1_.2876_or_0x4C.29)
	maxOpReturnPkScriptSize = 83
)

// Parsed proof represent semantically valid:
// - Bitcoin Header
// - Bitcoin Header hash
// - Bitcoin Transaction
// - Bitcoin Transaction index in block
// - Non-empty OpReturnData
type ParsedProof struct {
	// keeping header hash to avoid recomputing it everytime
	BlockHash        types.BTCHeaderHashBytes
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
	return chainhash.DoubleHashH(c)
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

func ExtractOpReturnData(tx *btcutil.Tx) []byte {
	msgTx := tx.MsgTx()
	opReturnData := []byte{}

	for _, output := range msgTx.TxOut {
		pkScript := output.PkScript
		pkScriptLen := len(pkScript)
		// valid op return script will have at least 2 bytes
		// - fisrt byte should be OP_RETURN marker
		// - second byte should indicate how many bytes there are in opreturn script
		if pkScriptLen > 1 &&
			pkScriptLen <= maxOpReturnPkScriptSize &&
			pkScript[0] == txscript.OP_RETURN {

			// if this is OP_PUSHDATA1, we need to drop first 3 bytes as those are related
			// to script iteslf i.e OP_RETURN + OP_PUSHDATA1 + len of bytes
			if pkScript[1] == txscript.OP_PUSHDATA1 {
				opReturnData = append(opReturnData, pkScript[3:]...)
			} else {
				// this should be one of OP_DATAXX opcodes we drop first 2 bytes
				opReturnData = append(opReturnData, pkScript[2:]...)
			}
		}
	}

	return opReturnData
}

func parseTransaction(bytes []byte) (*btcutil.Tx, error) {
	tx, e := btcutil.NewTxFromBytes(bytes)

	if e != nil {
		return nil, e
	}

	e = blockchain.CheckTransactionSanity(tx)

	if e != nil {
		return nil, e
	}

	return tx, nil
}

// TODO define domain errors with nice error messages
// TODO add some tests for the proof validation
func ParseProof(
	btcTransaction []byte,
	transactionIndex uint32,
	merkleProof []byte,
	btcHeader []byte,
	powLimit *big.Int) (*ParsedProof, error) {
	tx, e := parseTransaction(btcTransaction)

	if e != nil {
		return nil, e
	}

	header := types.BTCHeaderBytes(btcHeader).ToBlockHeader()

	e = types.ValidateBTCHeader(header, powLimit)

	if e != nil {
		return nil, e
	}

	validProof := prove(tx, &header.MerkleRoot, merkleProof, transactionIndex)

	if !validProof {
		return nil, fmt.Errorf("header failed validation due to failed proof")
	}

	opReturnData := ExtractOpReturnData(tx)

	if len(opReturnData) == 0 {
		return nil, fmt.Errorf("provided transaction should provide op return data")
	}

	bh := header.BlockHash()
	parsedProof := &ParsedProof{
		BlockHash:        types.NewBTCHeaderHashBytesFromChainhash(&bh),
		Transaction:      tx,
		TransactionBytes: btcTransaction,
		TransactionIdx:   transactionIndex,
		OpReturnData:     opReturnData,
	}

	return parsedProof, nil
}
