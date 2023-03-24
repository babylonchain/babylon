package grandpa_types

import (
	"bytes"
	"fmt"

	substrateCodec "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	substraterpcclienttypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type BlockNumber U32

type Hash [32]byte

type Header struct {
	ParentHash     Hash        `json:"parentHash"`
	Number         BlockNumber `json:"number"`
	StateRoot      Hash        `json:"stateRoot"`
	ExtrinsicsRoot Hash        `json:"extrinsicsRoot"`
	Digest         Digest      `json:"digest"`
}

func (header *GrandpaHeader) Encode() ([]byte, error) {
	var b bytes.Buffer

	if err := substrateCodec.NewEncoder(b).Encode(header); err != nil {
		return b.Bytes(), err
	}
	return b.Bytes(), nil

}

func (header *GrandpaHeader) Decode(decoder substrateCodec.Decoder) error {
	var bn uint32
	if err := decoder.Decode(&bn); err != nil {
		return err
	}
	header.Number = substraterpcclienttypes.BlockNumber(bn)
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
