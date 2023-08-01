package types

import (
	"bytes"
	"fmt"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// msgToSignForVote returns the message for an EOTS signature
// The EOTS signature on a block will be (blockHeight || blockHash)
func msgToSignForVote(blockHeight uint64, blockHash []byte) []byte {
	return append(sdk.Uint64ToBigEndian(blockHeight), blockHash...)
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

func (e *Evidence) canonicalMsgToSign() []byte {
	return msgToSignForVote(e.BlockHeight, e.CanonicalLastCommitHash)
}

func (e *Evidence) forkMsgToSign() []byte {
	return msgToSignForVote(e.BlockHeight, e.ForkLastCommitHash)
}

func (e *Evidence) ValidateBasic() error {
	if e.ValBtcPk == nil {
		return fmt.Errorf("empty ValBtcPk")
	}
	if e.PubRand == nil {
		return fmt.Errorf("empty PubRand")
	}
	if len(e.CanonicalLastCommitHash) != 32 {
		return fmt.Errorf("malformed CanonicalLastCommitHash")
	}
	if len(e.ForkLastCommitHash) != 32 {
		return fmt.Errorf("malformed ForkLastCommitHash")
	}
	if e.ForkFinalitySig == nil {
		return fmt.Errorf("empty ValBtcPk")
	}
	return nil
}

func (e *Evidence) IsSlashable() bool {
	if err := e.ValidateBasic(); err != nil {
		return false
	}
	if e.CanonicalFinalitySig == nil {
		return false
	}
	return true
}

// ExtractBTCSK extracts the BTC SK given the data in the evidence
func (e *Evidence) ExtractBTCSK() (*btcec.PrivateKey, error) {
	if !e.IsSlashable() {
		return nil, fmt.Errorf("the evidence lacks some fields so does not allow extracting BTC SK")
	}
	btcPK, err := e.ValBtcPk.ToBTCPK()
	if err != nil {
		return nil, err
	}
	return eots.Extract(
		btcPK, e.PubRand.ToFieldVal(),
		e.canonicalMsgToSign(), e.CanonicalFinalitySig.ToModNScalar(), // msg and sig for canonical block
		e.forkMsgToSign(), e.ForkFinalitySig.ToModNScalar(), // msg and sig for fork block
	)
}
