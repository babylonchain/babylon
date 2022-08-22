package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// IsValid verifies the validity of PoP
// 1. verify(sig=bls_sig, pubkey=blsPubkey, msg=pop.ed25519_sig)?
// 2. verify(sig=pop.ed25519_sig, pubkey=valPubkey, msg=blsPubkey)?
// BLS_pk ?= decrypt(key = Ed25519_pk, data = decrypt(key = BLS_pk, data = PoP))
func (pop ProofOfPossession) IsValid(blsPubKey bls12381.PublicKey) bool {
	ok, _ := bls12381.Verify(*pop.BlsSig, blsPubKey, pop.Ed25519Sig)
	if !ok {
		return false
	}
	ed25519PK := ed25519.PubKey(pop.Ed25519Pk)
	return ed25519PK.VerifySignature(blsPubKey, pop.Ed25519Sig)
}
