package types

import (
	"bytes"
	"github.com/btcsuite/btcd/wire"
)

func BytesToBtcdHeader(headerBytes *BTCHeader) (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}
	// The Deserialize method expects an io.Reader instance
	reader := bytes.NewReader(headerBytes.Header)
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
