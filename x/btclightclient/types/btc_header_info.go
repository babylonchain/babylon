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

func (hi *BTCHeaderInfo) HasParent(parent *BTCHeaderInfo) bool {
	return hi.Header.HasParent(parent.Header)
}

func (hi *BTCHeaderInfo) Eq(other *BTCHeaderInfo) bool {
	return hi.Hash.Eq(other.Hash)
}

func (hi BTCHeaderInfo) Validate() error {
	return nil
}
