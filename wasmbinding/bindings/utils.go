package bindings

import (
	lcTypes "github.com/babylonchain/babylon/x/btclightclient/types"
)

// translate BTCHeaderInfo to BtcBlockHeaderInfo
func AsBtcBlockHeaderInfo(info *lcTypes.BTCHeaderInfo) *BtcBlockHeaderInfo {
	if info == nil {
		return nil
	}

	header := info.Header.ToBlockHeader()
	return &BtcBlockHeaderInfo{
		Header: &BtcBlockHeader{
			Version: header.Version,
			Time:    uint32(header.Timestamp.Unix()),
			Bits:    header.Bits,
			Nonce:   header.Nonce,
			// for compatibility with all btc infra we are returning the hex encoded bytes
			// in reversed order
			MerkleRoot:    header.MerkleRoot.String(),
			PrevBlockhash: header.PrevBlock.String(),
		},
		Height: info.Height,
	}
}
