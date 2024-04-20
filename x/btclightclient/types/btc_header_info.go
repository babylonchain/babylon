package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
)

func NewBTCHeaderInfo(header *bbn.BTCHeaderBytes, headerHash *bbn.BTCHeaderHashBytes, height uint64, work *sdkmath.Uint) *BTCHeaderInfo {
	return &BTCHeaderInfo{
		Header: header,
		Hash:   headerHash,
		Height: height,
		Work:   work,
	}
}

func (m *BTCHeaderInfo) HasParent(parent *BTCHeaderInfo) bool {
	return m.Header.HasParent(parent.Header)
}

func (m *BTCHeaderInfo) Eq(other *BTCHeaderInfo) bool {
	return m.Hash.Eq(other.Hash)
}

// Validate verifies that the information inside the BTCHeaderInfo is valid.
func (m *BTCHeaderInfo) Validate() error {
	if m.Header == nil {
		return errors.New("header is nil")
	}
	if m.Hash == nil {
		return errors.New("hash is nil")
	}
	if m.Work == nil {
		return errors.New("work is nil")
	}

	if m.Work.IsZero() {
		return errors.New("work is zero")
	}

	btcHeader, err := bbn.NewBlockHeader(*m.Header)
	if err != nil {
		return err
	}

	blkHash := btcHeader.BlockHash()
	headerHash := bbn.NewBTCHeaderHashBytesFromChainhash(&blkHash)
	if !m.Hash.Eq(&headerHash) {
		return fmt.Errorf("BTC header hash is not equal to generated hash from header %s != %s", m.Hash, &headerHash)
	}

	return nil
}

func NewBTCHeaderInfoResponse(header *bbn.BTCHeaderBytes, headerHash *bbn.BTCHeaderHashBytes, height uint64, work *sdkmath.Uint) *BTCHeaderInfoResponse {
	return &BTCHeaderInfoResponse{
		HeaderHex: hex.EncodeToString(*header),
		HashHex:   hex.EncodeToString(*headerHash),
		Height:    height,
		Work:      *work,
	}
}
