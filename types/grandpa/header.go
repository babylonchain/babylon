package grandpa_types

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
)

type Header struct {
	ParentHash     Hash        `json:"parentHash"`
	Number         BlockNumber `json:"number"`
	StateRoot      Hash        `json:"stateRoot"`
	ExtrinsicsRoot Hash        `json:"extrinsicsRoot"`
	Digest         Digest      `json:"digest"`
}

// Encode encodes `value` with the scale codec with passed EncoderOptions, returning []byte
func (header *Header) Encode(encoder Encoder) ([]byte, error) {
	var buffer = bytes.Buffer{}
	err := encoder.Encode(header)
	if err != nil {
		return buffer.Bytes(), err
	}
	return buffer.Bytes(), nil
}

func DecodeHeader(decoder *Decoder) (*Header, error) {
	h := &Header{}
	if err := decoder.Decode(&h); err != nil {
		return h, err
	}
	return h, nil
	/*
		return nil, nil
		var bn uint32
		err = decoder.Decode(&bn)
		if err != nil {
			return err
		}
		header.Number = BlockNumber(bn)
		err = decoder.Decode(&header.ParentHash)
		if err != nil {
			return err
		}
		err = decoder.Decode(&header.StateRoot)
		if err != nil {
			return err
		}
		err = decoder.Decode(&header.ExtrinsicsRoot)
		if err != nil {
			return err
		}
		err = decoder.Decode(&header.Digest)
		if err != nil {
			return err
		}
		_, err = decoder.ReadOneByte()
		if err != nil {
			return fmt.Errorf("unexpected data after decoding header") // why??
		}
		return nil
	*/
}

type BlockNumber U32

// UnmarshalJSON fills BlockNumber with the JSON encoded byte array given by bz
func (b *BlockNumber) UnmarshalJSON(bz []byte) error {
	var tmp string
	if err := json.Unmarshal(bz, &tmp); err != nil {
		return err
	}

	s := strings.TrimPrefix(tmp, "0x")

	p, err := strconv.ParseUint(s, 16, 32)
	*b = BlockNumber(p)
	return err
}

// MarshalJSON returns a JSON encoded byte array of BlockNumber
func (b BlockNumber) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(b), 16)
	return json.Marshal(s)
}

// Encode implements encoding for BlockNumber, which just unwraps the bytes of BlockNumber
func (b BlockNumber) Encode(encoder Encoder) error {
	return encoder.EncodeUintCompact(*big.NewInt(0).SetUint64(uint64(b)))
}

// Decode implements decoding for BlockNumber, which just wraps the bytes in BlockNumber
func (b *BlockNumber) Decode(decoder Decoder) error {
	u, err := decoder.DecodeUintCompact()
	if err != nil {
		return err
	}
	*b = BlockNumber(u.Uint64())
	return err
}
