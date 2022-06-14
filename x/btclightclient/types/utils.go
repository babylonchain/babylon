package types

import (
	"bytes"
	"github.com/btcsuite/btcd/wire"
)

func BytesToBtcdHeader(headerBytes *BTCHeaderBytes) (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}
	// The Deserialize method expects an io.Reader instance
	reader := bytes.NewReader(headerBytes.HeaderBytes)
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
		PrevBlock:  []byte(btcdHeader.PrevBlock.String()),
		MerkleRoot: []byte(btcdHeader.MerkleRoot.String()),
		Bits:       btcdHeader.Bits,
		Nonce:      btcdHeader.Nonce,
		Hash:       []byte(btcdHeader.BlockHash().String()),
	}
}
