package types

import (
	bbl "github.com/babylonchain/babylon/types"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Ensure that MsgInsertHeader implements all functions of the Msg interface
var _ sdk.Msg = (*MsgInsertHeader)(nil)

func NewMsgInsertHeader(signer sdk.AccAddress, headerHex string) (*MsgInsertHeader, error) {
	headerBytes, err := bbl.NewBTCHeaderBytesFromHex(headerHex)
	if err != nil {
		return nil, err
	}
	return &MsgInsertHeader{Signer: signer.String(), Header: &headerBytes}, nil
}

func (msg *MsgInsertHeader) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return err
	}

	return msg.ValidateHeader(btcchaincfg.MainNetParams.PowLimit)
}

func (msg *MsgInsertHeader) ValidateHeader(powLimit *big.Int) error {
	header, err := msg.Header.ToBlockHeader()
	if err != nil {
		return err
	}
	return bbl.ValidateHeader(header, powLimit)
}

func (msg *MsgInsertHeader) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		// Panic, since the GetSigners method is called after ValidateBasic
		// which performs the same check.
		panic(err)
	}

	return []sdk.AccAddress{signer}
}
