package types

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type BIP340PubKey []byte

const BIP340PubKeyLen = schnorr.PubKeyBytesLen

func NewBIP340PubKey(data []byte) (BIP340PubKey, error) {
	var pk BIP340PubKey
	err := pk.Unmarshal(data)
	return pk, err
}

func NewBIP340PubKeyFromBTCPK(btcPK *btcec.PublicKey) BIP340PubKey {
	pkBytes := schnorr.SerializePubKey(btcPK)
	return BIP340PubKey(pkBytes)
}

func (pk BIP340PubKey) ToBTCPK() (*btcec.PublicKey, error) {
	return schnorr.ParsePubKey(pk)
}

func (pk BIP340PubKey) Size() int {
	return len(pk.MustMarshal())
}

func (pk BIP340PubKey) Marshal() ([]byte, error) {
	return pk, nil
}

func (pk BIP340PubKey) MustMarshal() []byte {
	pkBytes, err := pk.Marshal()
	if err != nil {
		panic(err)
	}
	return pkBytes
}

func (pk BIP340PubKey) MarshalTo(data []byte) (int, error) {
	bz, err := pk.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (pk *BIP340PubKey) Unmarshal(data []byte) error {
	newPK := BIP340PubKey(data)

	// ensure that the bytes can be transformed to a *btcec.PublicKey object
	// this includes all format checks
	_, err := newPK.ToBTCPK()
	if err != nil {
		return errors.New("bytes cannot be converted to a *btcec.PublicKey object")
	}

	*pk = data
	return nil
}
