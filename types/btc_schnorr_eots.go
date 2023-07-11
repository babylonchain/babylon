package types

import (
	"bytes"
	"encoding/hex"
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

func NewSchnorrEOTSSigFromHex(sigHex string) (*SchnorrEOTSSig, error) {
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, err
	}
	return NewSchnorrEOTSSig(sigBytes)
}

func NewSchnorrEOTSSigFromModNScalar(s *btcec.ModNScalar) *SchnorrEOTSSig {
	prBytes := s.Bytes()
	sig := SchnorrEOTSSig(prBytes[:])
	return &sig
}

func (sig SchnorrEOTSSig) ToModNScalar() *btcec.ModNScalar {
	var s btcec.ModNScalar
	s.SetByteSlice(sig)
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

func (sig *SchnorrEOTSSig) Equals(sig2 *SchnorrEOTSSig) bool {
	return bytes.Equal(sig.MustMarshal(), sig2.MustMarshal())
}

func (sig *SchnorrEOTSSig) ToHexStr() string {
	sigBytes := sig.MustMarshal()
	return hex.EncodeToString(sigBytes)
}
