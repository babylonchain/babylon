package types

import (
	"bytes"
	"encoding/hex"
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

func NewBTCTxFromBytes(txBytes []byte) (*wire.MsgTx, error) {
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(txBytes)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return nil, err
	}

	return &msgTx, nil
}

func NewBTCTxFromHex(txHex string) (*wire.MsgTx, []byte, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := NewBTCTxFromBytes(txBytes)

	if err != nil {
		return nil, nil, err
	}

	return parsed, txBytes, nil
}

func SerializeBTCTx(tx *wire.MsgTx) ([]byte, error) {
	var txBuf bytes.Buffer
	if err := tx.Serialize(&txBuf); err != nil {
		return nil, err
	}
	return txBuf.Bytes(), nil
}

func GetOutputIdxInBTCTx(tx *wire.MsgTx, output *wire.TxOut) (uint32, error) {
	for i, txOut := range tx.TxOut {
		if bytes.Equal(txOut.PkScript, output.PkScript) && txOut.Value == output.Value {
			return uint32(i), nil
		}
	}

	return 0, fmt.Errorf("output not found")
}
