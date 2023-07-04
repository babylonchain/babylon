package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ib *IndexedBlock) Equal(ib2 *IndexedBlock) bool {
	if !bytes.Equal(ib.LastCommitHash, ib2.LastCommitHash) {
		return false
	}
	if ib.Height != ib2.Height {
		return false
	}
	// NOTE: we don't compare finalisation status here
	return true
}

func (ib *IndexedBlock) MsgToSign() []byte {
	return MsgToSignForVote(ib.Height, ib.LastCommitHash)
}

// MsgToSignForVote returns the message for an EOTS signature
// The EOTS signature on a block will be (blockHeight || blockHash)
func MsgToSignForVote(blockHeight uint64, blockHash []byte) []byte {
	msgToSign := []byte{}
	msgToSign = append(msgToSign, sdk.Uint64ToBigEndian(blockHeight)...)
	msgToSign = append(msgToSign, blockHash...)
	return msgToSign
}
