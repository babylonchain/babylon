package btclightclient

import (
	"bytes"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/wire"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BytesToBtcdHeader(headerBytes *types.BTCHeaderBytes) (*wire.BlockHeader, error) {
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

func BtcdHeaderToBTCBlockHeader(btcdHeader *wire.BlockHeader) *types.BTCBlockHeader {
	return &types.BTCBlockHeader{
		Version:    btcdHeader.Version,
		PrevBlock:  btcdHeader.PrevBlock.String(),
		MerkleRoot: btcdHeader.MerkleRoot.String(),
		Time:       timestamppb.New(btcdHeader.Timestamp),
		Bits:       btcdHeader.Bits,
		Nonce:      btcdHeader.Nonce,
		Hash:       btcdHeader.BlockHash().String(),
	}
}
