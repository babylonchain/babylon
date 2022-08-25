package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// IsValid verifies the validity of PoP
// 1. verify(sig=bls_sig, pubkey=blsPubkey, msg=pop.ed25519_sig)?
// 2. verify(sig=pop.ed25519_sig, pubkey=valPubkey, msg=blsPubkey)?
// BLS_pk ?= decrypt(key = Ed25519_pk, data = decrypt(key = BLS_pk, data = PoP))
func (pop ProofOfPossession) IsValid(blsPubkey bls12381.PublicKey, valPubkey cryptotypes.PubKey) bool {
	ok, _ := bls12381.Verify(*pop.BlsSig, blsPubkey, pop.Ed25519Sig)
	if !ok {
		return false
	}
	ed25519PK := ed25519.PubKey(valPubkey.Bytes())
	return ed25519PK.VerifySignature(blsPubkey, pop.Ed25519Sig)
}
