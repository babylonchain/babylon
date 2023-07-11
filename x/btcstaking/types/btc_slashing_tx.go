package types

import (
	"bytes"
	"encoding/hex"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
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

func NewBTCSlashingTxFromHex(txHex string) (*BTCSlashingTx, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}
	var tx BTCSlashingTx
	if err := tx.Unmarshal(txBytes); err != nil {
		return nil, err
	}
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
	*tx = data

	// ensure data can be decoded to a tx
	if _, err := tx.ToMsgTx(); err != nil {
		return err
	}

	return nil
}

func (tx *BTCSlashingTx) Size() int {
	return len(tx.MustMarshal())
}

func (tx *BTCSlashingTx) ToHexStr() string {
	txBytes := tx.MustMarshal()
	return hex.EncodeToString(txBytes)
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

// Sign generates a signature on the slashing tx signed by staker, validator or jury
func (tx *BTCSlashingTx) Sign(stakingMsgTx *wire.MsgTx, stakingScript []byte, sk *btcec.PrivateKey, net *chaincfg.Params) (*bbn.BIP340Signature, error) {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	schnorrSig, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
		msgTx,
		stakingMsgTx,
		sk,
		stakingScript,
		net,
	)
	if err != nil {
		return nil, err
	}
	delegatorSig := bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
	return &delegatorSig, nil
}

// VerifySignature verifies a signature on the slashing tx signed by staker, validator or jury
func (tx *BTCSlashingTx) VerifySignature(stakingPkScript []byte, stakingAmount int64, stakingScript []byte, pk *btcec.PublicKey, sig *bbn.BIP340Signature) error {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return err
	}
	return btcstaking.VerifyTransactionSigWithOutputData(
		msgTx,
		stakingPkScript,
		stakingAmount,
		stakingScript,
		pk,
		*sig,
	)
}
