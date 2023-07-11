package types

import (
	"encoding/hex"
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// NewPoP generates a new proof of possession that sk_Babylon and sk_BTC are held by the same person
// a proof of possession contains two signatures:
// - pop.BabylonSig = sign(sk_Babylon, pk_BTC)
// - pop.BtcSig = sign(sk_BTC, pop.BabylonSig)
func NewPoP(babylonSK cryptotypes.PrivKey, btcSK *btcec.PrivateKey) (*ProofOfPossession, error) {
	pop := ProofOfPossession{}

	// generate pop.BabylonSig = sign(sk_Babylon, pk_BTC)
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		return nil, err
	}
	pop.BabylonSig = babylonSig

	// generate pop.BtcSig = sign(sk_BTC, pop.BabylonSig)
	// NOTE: *schnorr.Sign has to take the hash of the message.
	// So we have to hash babylonSig before signing
	babylonSigHash := tmhash.Sum(pop.BabylonSig)
	btcSig, err := schnorr.Sign(btcSK, babylonSigHash)
	if err != nil {
		return nil, err
	}
	bip340Sig := bbn.NewBIP340SignatureFromBTCSig(btcSig)
	pop.BtcSig = &bip340Sig

	return &pop, nil
}

func NewPoPFromHex(popHex string) (*ProofOfPossession, error) {
	popBytes, err := hex.DecodeString(popHex)
	if err != nil {
		return nil, err
	}
	var pop ProofOfPossession
	if err := pop.Unmarshal(popBytes); err != nil {
		return nil, err
	}
	return &pop, nil
}

func (pop *ProofOfPossession) ToHexStr() (string, error) {
	popBytes, err := pop.Marshal()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(popBytes), nil
}

// Verify verifies the validity of PoP
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=pop.BabylonSig)?
// 2. verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
func (pop *ProofOfPossession) Verify(babylonPK cryptotypes.PubKey, bip340PK *bbn.BIP340PubKey) error {
	// rule 1: verify(sig=sig_btc, pubkey=pk_btc, msg=pop.BabylonSig)?
	btcSig, err := pop.BtcSig.ToBTCSig()
	if err != nil {
		return err
	}
	btcPK, err := bip340PK.ToBTCPK()
	if err != nil {
		return err
	}
	// NOTE: btcSig.Verify has to take hash of the message.
	// So we have to hash babylonSig before verifying the signature
	babylonSigHash := tmhash.Sum(pop.BabylonSig)
	if !btcSig.Verify(babylonSigHash, btcPK) {
		return fmt.Errorf("failed to verify babylonSig")
	}

	// rule 2: verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
	if !babylonPK.VerifySignature(*bip340PK, pop.BabylonSig) {
		return fmt.Errorf("failed to verify pop.BabylonSig")
	}

	return nil
}
