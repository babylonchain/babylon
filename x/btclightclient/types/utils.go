package types

import (
	"bytes"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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

func BtcdHeaderToBytes(header *wire.BlockHeader) *BTCHeaderBytes {
	var buf bytes.Buffer
	header.Serialize(&buf)

	return &BTCHeaderBytes{HeaderBytes: buf.Bytes()}
}

func BytesToChainhash(hashBytes []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(hashBytes)
}

func ChainhashToBytes(hash *chainhash.Hash) []byte {
	return hash.CloneBytes()
}
