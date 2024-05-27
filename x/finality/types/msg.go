package types

import (
	fmt "fmt"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgAddFinalitySig{}
	_ sdk.Msg = &MsgCommitPubRandList{}
)

func (m *MsgAddFinalitySig) MsgToSign() []byte {
	return msgToSignForVote(m.BlockHeight, m.BlockAppHash)
}

// VerifyFinalitySig verifies the finality signature message w.r.t. the
// public randomness commitment. The verification includes
// - verifying the proof of inclusion of the given public randomness
// - verifying the finality signature w.r.t. the given block height/hash
func VerifyFinalitySig(m *MsgAddFinalitySig, prCommit *PubRandCommit) error {
	// verify the index of the public randomness
	heightOfProof := prCommit.StartHeight + uint64(m.Proof.Index)
	if m.BlockHeight != heightOfProof {
		return ErrInvalidFinalitySig.Wrapf("the inclusion proof (for height %d) does not correspond to the given height (%d) in the message", heightOfProof, m.BlockHeight)
	}
	// verify the total number of randomness is same as in the commit
	if uint64(m.Proof.Total) != prCommit.NumPubRand {
		return ErrInvalidFinalitySig.Wrapf("the total number of public randomnesses in the proof (%d) does not match the number of public randomnesses committed (%d)", m.Proof.Total, prCommit.NumPubRand)
	}
	// verify the proof of inclusion for this public randomness
	unwrappedProof, err := merkle.ProofFromProto(m.Proof)
	if err != nil {
		return ErrInvalidFinalitySig.Wrapf("failed to unwrap proof: %v", err)
	}
	if err := unwrappedProof.Verify(prCommit.Commitment, *m.PubRand); err != nil {
		return ErrInvalidFinalitySig.Wrapf("the inclusion proof of the public randomness is invalid: %v", err)
	}

	// public randomness is good, verify finality signature
	msgToSign := m.MsgToSign()
	pk, err := m.FpBtcPk.ToBTCPK()
	if err != nil {
		return err
	}
	return eots.Verify(pk, m.PubRand.ToFieldVal(), msgToSign, m.FinalitySig.ToModNScalar())
}

// HashToSign returns a 32-byte hash of (start_height || num_pub_rand || commitment)
// The signature in MsgCommitPubRandList will be on this hash
func (m *MsgCommitPubRandList) HashToSign() ([]byte, error) {
	hasher := tmhash.New()
	if _, err := hasher.Write(sdk.Uint64ToBigEndian(m.StartHeight)); err != nil {
		return nil, err
	}
	if _, err := hasher.Write(sdk.Uint64ToBigEndian(m.NumPubRand)); err != nil {
		return nil, err
	}
	if _, err := hasher.Write(m.Commitment); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func (m *MsgCommitPubRandList) VerifySig() error {
	msgHash, err := m.HashToSign()
	if err != nil {
		return err
	}
	pk, err := m.FpBtcPk.ToBTCPK()
	if err != nil {
		return err
	}
	if m.Sig == nil {
		return fmt.Errorf("empty signature")
	}
	schnorrSig, err := m.Sig.ToBTCSig()
	if err != nil {
		return err
	}
	if !schnorrSig.Verify(msgHash, pk) {
		return fmt.Errorf("failed to verify signature")
	}
	return nil
}
