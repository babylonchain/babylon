package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

type SchnorrEOTSSig []byte

const SchnorrEOTSSigLen = 32

func NewSchnorrEOTSSig(data []byte) (*SchnorrEOTSSig, error) {
	var sig SchnorrEOTSSig
	err := sig.Unmarshal(data)
	return &sig, err
}

func NewSchnorrEOTSSigFromModNScalar(r *btcec.ModNScalar) *SchnorrEOTSSig {
	prBytes := r.Bytes()
	sig := SchnorrEOTSSig(prBytes[:])
	return &sig
}

func (sig SchnorrEOTSSig) ToModNScalar() *btcec.ModNScalar {
	var s btcec.ModNScalar
	s.PutBytesUnchecked(sig)
	return &s
}

func (sig SchnorrEOTSSig) Size() int {
	return len(sig.MustMarshal())
}

func (sig SchnorrEOTSSig) Marshal() ([]byte, error) {
	return sig, nil
}

func (sig SchnorrEOTSSig) MustMarshal() []byte {
	prBytes, err := sig.Marshal()
	if err != nil {
		panic(err)
	}
	return prBytes
}

func (sig SchnorrEOTSSig) MarshalTo(data []byte) (int, error) {
	bz, err := sig.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (sig *SchnorrEOTSSig) Unmarshal(data []byte) error {
	if len(data) != SchnorrEOTSSigLen {
		return fmt.Errorf("invalid data length")
	}
	*sig = data
	return nil
}
