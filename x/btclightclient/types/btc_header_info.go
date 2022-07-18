package types

import (
	bbl "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewBTCHeaderInfo(header *bbl.BTCHeaderBytes, headerHash *bbl.BTCHeaderHashBytes, height uint64, work *sdk.Uint) *BTCHeaderInfo {
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

func (m BTCHeaderInfo) Validate() error {
	return nil
}
