package eots

import (
	"errors"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type MasterSecretRand struct {
	k *hdkeychain.ExtendedKey
}
type MasterPublicRand struct {
	k *hdkeychain.ExtendedKey
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
	var (
		seed [32]byte
		err  error
	)
	if _, err := io.ReadFull(randSource, seed[:]); err != nil {
		return nil, nil, err
	}
	// generate new master key pair
	var masterSK *hdkeychain.ExtendedKey
	for {
		masterSK, err = hdkeychain.NewMaster(seed[:], &chaincfg.MainNetParams)
		// if all good, use this master SK
		if err == nil {
			break
		}
		// NOTE: There is an extremely small chance (< 1 in 2^127) the provided seed
		// will derive to an unusable secret key.  The ErrUnusableSeed error will be
		// returned if this should occur. We need to try to generate a new master SK
		// again
		if errors.Is(err, hdkeychain.ErrUnusableSeed) {
			continue
		}
		// some other unrecoverable error, return error
		return nil, nil, err
	}

	masterPK, err := masterSK.Neuter()
	if err != nil {
		return nil, nil, err
	}

	return &MasterSecretRand{masterSK}, &MasterPublicRand{masterPK}, nil
}

func (msr *MasterSecretRand) Validate() error {
	if !msr.k.IsPrivate() {
		return fmt.Errorf("underlying key is not a private key")
	}
	return nil
}

func NewMasterSecretRandFromBase58(s string) (*MasterSecretRand, error) {
	k, err := hdkeychain.NewKeyFromString(s)
	if err != nil {
		return nil, err
	}
	if !k.IsPrivate() {
		return nil, fmt.Errorf("the given string does not correspond to a secret key")
	}
	return &MasterSecretRand{k}, nil
}

func NewMasterSecretRand(b []byte) (*MasterSecretRand, error) {
	return NewMasterSecretRandFromBase58(string(b))
}

func (msr *MasterSecretRand) MasterPubicRand() (*MasterPublicRand, error) {
	masterPK, err := msr.k.Neuter()
	if err != nil {
		return nil, err
	}
	return &MasterPublicRand{masterPK}, nil
}

// TODO: extend to support uint64
func (msr *MasterSecretRand) DeriveRandPair(height uint32) (*PrivateRand, *PublicRand, error) {
	// get child SK, then child SK in BTC format, and finally private randomness
	childSK, err := msr.k.Derive(height)
	if err != nil {
		return nil, nil, err
	}
	childBTCSK, err := childSK.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}
	privRand := &childBTCSK.Key

	// get child PK in BTC format, and then public randomness
	childBTCPK := childBTCSK.PubKey()
	var j secp256k1.JacobianPoint
	childBTCPK.AsJacobian(&j)
	pubRand := &j.X

	return privRand, pubRand, nil
}

func (msr *MasterSecretRand) MarshalBase58() string {
	return msr.k.String()
}

func (msr *MasterSecretRand) Marshal() []byte {
	return []byte(msr.MarshalBase58())
}

func NewMasterPublicRandFromBase58(s string) (*MasterPublicRand, error) {
	k, err := hdkeychain.NewKeyFromString(s)
	if err != nil {
		return nil, err
	}
	if k.IsPrivate() {
		return nil, fmt.Errorf("the given string does not correspond to a public key")
	}
	return &MasterPublicRand{k}, nil
}

func NewMasterPublicRand(b []byte) (*MasterPublicRand, error) {
	return NewMasterPublicRandFromBase58(string(b))
}

func (mpr *MasterPublicRand) Validate() error {
	if mpr.k.IsPrivate() {
		return fmt.Errorf("underlying key is not a public key")
	}
	return nil
}

func (mpr *MasterPublicRand) DerivePubRand(height uint32) (*PublicRand, error) {
	childPK, err := mpr.k.Derive(height)
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

func (mpr *MasterPublicRand) MarshalBase58() string {
	return mpr.k.String()
}

func (mpr *MasterPublicRand) Marshal() []byte {
	return []byte(mpr.MarshalBase58())
}
