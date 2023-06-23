package types

import (
	"bytes"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

type BTCSlashingTx []byte

// TODO: constructors, verification rules, and util functions

func NewBTCSlashingTxFromMsgTx(msgTx *wire.MsgTx) (*BTCSlashingTx, error) {
	var buf bytes.Buffer
	err := msgTx.Serialize(&buf)
	if err != nil {
		return nil, err
	}

	tx := BTCSlashingTx(buf.Bytes())
	return &tx, nil
}

func (tx BTCSlashingTx) Marshal() ([]byte, error) {
	return tx, nil
}

func (tx BTCSlashingTx) MustMarshal() []byte {
	txBytes, err := tx.Marshal()
	if err != nil {
		panic(err)
	}
	return txBytes
}

func (tx BTCSlashingTx) MarshalTo(data []byte) (int, error) {
	bz, err := tx.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (tx *BTCSlashingTx) Unmarshal(data []byte) error {
	// TODO: verifications

	*tx = data
	return nil
}

func (tx *BTCSlashingTx) Size() int {
	return len(tx.MustMarshal())
}

func (tx *BTCSlashingTx) ToMsgTx() (*wire.MsgTx, error) {
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(*tx)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return nil, err
	}
	return &msgTx, nil
}

func (tx *BTCSlashingTx) Validate(net *chaincfg.Params, slashingAddress string) error {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return err
	}
	decodedAddr, err := btcutil.DecodeAddress(slashingAddress, net)
	if err != nil {
		return err
	}
	return btcstaking.CheckSlashingTx(msgTx, decodedAddr)
}
