package types

import (
	"bytes"

	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// msgToSignForVote returns the message for an EOTS signature
// The EOTS signature on a block will be (blockHeight || blockHash)
func msgToSignForVote(blockHeight uint64, blockHash []byte) []byte {
	msgToSign := []byte{}
	msgToSign = append(msgToSign, sdk.Uint64ToBigEndian(blockHeight)...)
	msgToSign = append(msgToSign, blockHash...)
	return msgToSign
}

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
	return msgToSignForVote(ib.Height, ib.LastCommitHash)
}

func (e *Evidence) MsgToSign() []byte {
	return msgToSignForVote(e.BlockHeight, e.BlockLastCommitHash)
}

// ExtractBTCSK extracts the BTC SK given the canonical block, public randomness
// and EOTS signature on the canonical block
func (e *Evidence) ExtractBTCSK(indexedBlock *IndexedBlock, pubRand *bbn.SchnorrPubRand, sig *bbn.SchnorrEOTSSig) (*btcec.PrivateKey, error) {
	btcPK, err := e.ValBtcPk.ToBTCPK()
	if err != nil {
		return nil, err
	}
	return eots.Extract(
		btcPK, pubRand.ToFieldVal(),
		indexedBlock.MsgToSign(), sig.ToModNScalar(), // msg and sig for canonical block
		e.MsgToSign(), e.FinalitySig.ToModNScalar(), // msg and sig for fork block
	)
}
