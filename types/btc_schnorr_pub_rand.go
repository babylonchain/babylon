package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

type SchnorrPubRand []byte

const SchnorrPubRandLen = 32

func NewSchnorrPubRand(data []byte) (*SchnorrPubRand, error) {
	var pr SchnorrPubRand
	err := pr.Unmarshal(data)
	return &pr, err
}

func NewSchnorrPubRandFromFieldVal(r *btcec.FieldVal) *SchnorrPubRand {
	prBytes := r.Bytes()
	pr := SchnorrPubRand(prBytes[:])
	return &pr
}

func (pr SchnorrPubRand) ToFieldVal() *btcec.FieldVal {
	var r btcec.FieldVal
	r.SetByteSlice(pr)
	return &r
}

func (pr SchnorrPubRand) Size() int {
	return len(pr.MustMarshal())
}

func (pr SchnorrPubRand) Marshal() ([]byte, error) {
	return pr, nil
}

func (pr SchnorrPubRand) MustMarshal() []byte {
	prBytes, err := pr.Marshal()
	if err != nil {
		panic(err)
	}
	return prBytes
}

func (pr SchnorrPubRand) MarshalTo(data []byte) (int, error) {
	bz, err := pr.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (pr *SchnorrPubRand) Unmarshal(data []byte) error {
	if len(data) != SchnorrPubRandLen {
		return fmt.Errorf("invalid data length")
	}
	*pr = data
	return nil
}
