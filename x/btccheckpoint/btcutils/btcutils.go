package btcutils

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"time"

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
// - Bitcoin Transaction
// - Non-empty OpReturnData
type ParsedProof struct {
	BlockHeader  wire.BlockHeader
	Transaction  *btcutil.Tx
	OpReturnData []byte
}

func readBlockHeader(headerBytes []byte) (*wire.BlockHeader, error) {
	header := &wire.BlockHeader{}

	reader := bytes.NewReader(headerBytes)

	e := header.Deserialize(reader)

	if e != nil {
		return nil, e
	}

	return header, nil
}

// TODO copy of the validation done in btc light client at some point it would
// be nice to move it to some commong btc module
func validateHeader(header *wire.BlockHeader, powLimit *big.Int) error {
	// Perform the checks that checkBlockHeaderSanity of btcd does
	// https://github.com/btcsuite/btcd/blob/master/blockchain/validate.go#L430
	// We skip the "timestamp should not be 2 hours into the future" check
	// since this might introduce undeterministic behavior

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

// Concatenates and double hashes two provided inputs
func hashConcat(a []byte, b []byte) chainhash.Hash {
	c := []byte{}
	c = append(c, a...)
	c = append(c, b...)
	return chainhash.HashH(c)
}

// proof logic copied from:
// https://github.com/summa-tx/bitcoin-spv/blob/fb2a61e7a941d421ae833789d97ed10d2ad79cfe/golang/btcspv/bitcoin_spv.go#L498
// main reason for not bringing library in, is that we already use btcd
// bitcoin primitives and this library defines their own which could lead
// to some mixups
func verifyProof(proof []byte, index uint32) bool {
	var current chainhash.Hash
	idx := index
	proofLength := len(proof)

	if proofLength%32 != 0 {
		return false
	}

	if proofLength == 32 {
		return true
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

// Prove checks the validity of a merkle proof
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

	return verifyProof(proof, index)
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

	header, e := readBlockHeader(btcHeader)

	if e != nil {
		return nil, e
	}

	e = validateHeader(header, powLimit)

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
		BlockHeader:  *header,
		Transaction:  tx,
		OpReturnData: opReturnData,
	}

	return parsedProof, nil
}
