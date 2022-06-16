package types

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/wire"
)

func BytesToBtcdHeader(headerBytes *BTCHeader) (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}

	// TODO: remove
	headerString, _ := hex.DecodeString("00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22")
	headeroo := []byte(headerString)

	// The Deserialize method expects an io.Reader instance
	// reader := bytes.NewReader(headerBytes.Header)
	reader := bytes.NewReader(headeroo)
	// Decode the header bytes
	err := header.Deserialize(reader)
	// There was a parsing error
	if err != nil {
		return nil, err
	}
	return header, nil
}

func BtcdHeaderToBTCBlockHeader(btcdHeader *wire.BlockHeader) *BTCBlockHeader {
	return &BTCBlockHeader{
		Version:    btcdHeader.Version,
		PrevBlock:  BlockHash(btcdHeader.PrevBlock.String()),
		MerkleRoot: []byte(btcdHeader.MerkleRoot.String()),
		Bits:       btcdHeader.Bits,
		Nonce:      btcdHeader.Nonce,
		Hash:       BlockHash(btcdHeader.BlockHash().String()),
	}
}
