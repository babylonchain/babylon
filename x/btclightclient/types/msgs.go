package types

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Ensure that MsgInsertHeader implements all functions of the Msg interface
var _ sdk.Msg = (*MsgInsertHeader)(nil)

func NewMsgInsertHeader(signer sdk.AccAddress, header []byte) *MsgInsertHeader {
	headerBytes := &BTCHeader{Header: header}
	return &MsgInsertHeader{Signer: signer.String(), Header: headerBytes}
}

func (msg *MsgInsertHeader) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return err
	}

	header, err := BytesToBtcdHeader(msg.Header)
	if err != nil {
		return err
	}

	return ValidateHeaderAttributes(header)
}

func ValidateHeaderAttributes(header *wire.BlockHeader) error {
	// Perform the checks that checkBlockHeaderSanity of btcd does
	// https://github.com/btcsuite/btcd/blob/master/blockchain/validate.go#L430
	// We skip the "timestamp should not be 2 hours into the future" check
	// since this might introduce undeterministic behavior

	msgBlock := &wire.MsgBlock{Header: *header}
	block := btcutil.NewBlock(msgBlock)

	// The upper limit for the power to be spent
	// Use the one maintained by btcd
	powLimit := btcchaincfg.MainNetParams.PowLimit

	err := blockchain.CheckProofOfWork(block, powLimit)
	if err != nil {
		return err
	}

	if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
		str := fmt.Sprintf("block timestamp of %v has a higher "+
			"precision than one second", header.Timestamp)
		return errors.New(str)
	}

	return nil
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
