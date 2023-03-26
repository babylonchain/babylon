package grandpa_types

import (
	"bytes"
	"fmt"
)

type BlockNumber U32

type Header struct {
	ParentHash     Hash        `json:"parentHash"`
	Number         BlockNumber `json:"number"`
	StateRoot      Hash        `json:"stateRoot"`
	ExtrinsicsRoot Hash        `json:"extrinsicsRoot"`
	Digest         Digest      `json:"digest"`
}

func (header *Header) Encode() ([]byte, error) {
	var b = bytes.Buffer{}

	if err := NewEncoder(&b).Encode(header); err != nil {
		return b.Bytes(), err
	}
	return b.Bytes(), nil

}

func (header *Header) Decode(decoder Decoder) error {
	var bn uint32
	if err := decoder.Decode(&bn); err != nil {
		return err
	}
	header.Number = BlockNumber(bn)
	if err := decoder.Decode(&header.ParentHash); err != nil {
		return err
	}
	if err := decoder.Decode(&header.StateRoot); err != nil {
		return err
	}
	if err := decoder.Decode(&header.ExtrinsicsRoot); err != nil {
		return err
	}
	if err := decoder.Decode(&header.Digest); err != nil {
		return err
	}
	if _, err := decoder.ReadOneByte(); err == nil {
		return fmt.Errorf("unexpected data after decoding header")
	}
	return nil
}
