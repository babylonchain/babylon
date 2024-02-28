package eots

import (
	"io"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type MasterSecretRand struct {
	*hdkeychain.ExtendedKey
}
type MasterPublicRand struct {
	*hdkeychain.ExtendedKey
}

type PrivateRand = secp256k1.ModNScalar
type PublicRand = secp256k1.FieldVal

// RandGen returns the value to be used as random value when signing, and the associated public value.
func RandGen(randSource io.Reader) (*PrivateRand, *PublicRand, error) {
	pk, err := KeyGen(randSource)
	if err != nil {
		return nil, nil, err
	}
	var j secp256k1.JacobianPoint
	pk.PubKey().AsJacobian(&j)
	return &pk.Key, &j.X, nil
}

func NewMasterRandPair(randSource io.Reader) (*MasterSecretRand, *MasterPublicRand, error) {
	// get random seed
	var seed [32]byte
	if _, err := io.ReadFull(randSource, seed[:]); err != nil {
		return nil, nil, err
	}
	// generate new master key pair
	masterSK, err := hdkeychain.NewMaster(seed[:], &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, err
	}
	masterPK, err := masterSK.Neuter()
	if err != nil {
		return nil, nil, err
	}

	return &MasterSecretRand{masterSK}, &MasterPublicRand{masterPK}, nil
}

func (msr *MasterSecretRand) DeriveRandPair(height uint32) (*PrivateRand, *PublicRand, error) {
	// get child SK, then child SK in BTC format, and finally private randomness
	childSK, err := msr.Derive(height)
	if err != nil {
		return nil, nil, err
	}
	childBTCSK, err := childSK.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}
	privRand := &childBTCSK.Key

	// get child PK, then child PK in BTC format, and finally private randomness
	childPK, err := childSK.Neuter()
	if err != nil {
		return nil, nil, err
	}
	childBTCPK, err := childPK.ECPubKey()
	if err != nil {
		return nil, nil, err
	}
	var j secp256k1.JacobianPoint
	childBTCPK.AsJacobian(&j)
	pubRand := &j.X

	return privRand, pubRand, nil
}

func (mpr *MasterPublicRand) DerivePubRand(height uint32) (*PublicRand, error) {
	childPK, err := mpr.Derive(height)
	if err != nil {
		return nil, err
	}
	childBTCPK, err := childPK.ECPubKey()
	if err != nil {
		return nil, err
	}
	var j secp256k1.JacobianPoint
	childBTCPK.AsJacobian(&j)
	pubRand := &j.X
	return pubRand, nil
}
