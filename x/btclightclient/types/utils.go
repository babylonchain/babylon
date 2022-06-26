package types

import (
	"bytes"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// BytesToBtcdHeader parse header bytes into a BlockHeader instance
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

// BtcdHeaderToBytes gets a BlockHeader instance and returns the header bytes
func BtcdHeaderToBytes(header *wire.BlockHeader) *BTCHeaderBytes {
	var buf bytes.Buffer
	header.Serialize(&buf)

	return &BTCHeaderBytes{HeaderBytes: buf.Bytes()}
}

// BytesToChainhash gets hash bytes in reverse order and returns a Hash instance
func BytesToChainhash(hashBytes []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(hashBytes)
}

// ChainhashToBytes gets a Hash instance and returns bytes in reverse order
func ChainhashToBytes(hash *chainhash.Hash) []byte {
	return hash[:]
}
