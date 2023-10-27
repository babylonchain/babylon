package types

import (
	"encoding/hex"
	"fmt"
	"math/big"

	bbn "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = (*MsgInsertHeaders)(nil)

func NewMsgInsertHeaders(signer sdk.AccAddress, headersHex string) (*MsgInsertHeaders, error) {
	if len(headersHex) == 0 {
		return nil, fmt.Errorf("empty headers list")
	}

	decoded, err := hex.DecodeString(headersHex)

	if err != nil {
		return nil, err
	}

	if len(decoded)%bbn.BTCHeaderLen != 0 {
		return nil, fmt.Errorf("invalid length of encoded headers: %d", len(decoded))
	}
	numOfHeaders := len(decoded) / bbn.BTCHeaderLen
	headers := make([]bbn.BTCHeaderBytes, numOfHeaders)

	for i := 0; i < numOfHeaders; i++ {
		headerSlice := decoded[i*bbn.BTCHeaderLen : (i+1)*bbn.BTCHeaderLen]
		headerBytes, err := bbn.NewBTCHeaderBytesFromBytes(headerSlice)
		if err != nil {
			return nil, err
		}
		headers[i] = headerBytes
	}
	return &MsgInsertHeaders{Signer: signer.String(), Headers: headers}, nil
}

func (msg *MsgInsertHeaders) ValidateBasic() error {
	// This function validates stateless message elements
	// msg.Header is validated in ante-handler
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgInsertHeaders) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		// Panic, since the GetSigners method is called after ValidateBasic
		// which performs the same check.
		panic(err)
	}

	return []sdk.AccAddress{signer}
}

func (msg *MsgInsertHeaders) ValidateHeaders(powLimit *big.Int) error {
	// TOOD: Limit number of headers in message?
	for _, header := range msg.Headers {
		err := bbn.ValidateBTCHeader(header.ToBlockHeader(), powLimit)
		if err != nil {
			return err
		}
	}

	return nil
}
