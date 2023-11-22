package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/wire"
)

type PublicKeyInfo struct {
	StakerKey    *bbn.BIP340PubKey
	ValidatorKey *bbn.BIP340PubKey
	CovenantKey  *bbn.BIP340PubKey
}

func KeyDataFromScript(scriptData *btcstaking.StakingScriptData) *PublicKeyInfo {
	return &PublicKeyInfo{
		StakerKey:    bbn.NewBIP340PubKeyFromBTCPK(scriptData.StakerKey),
		ValidatorKey: bbn.NewBIP340PubKeyFromBTCPK(scriptData.ValidatorKey),
		CovenantKey:  bbn.NewBIP340PubKeyFromBTCPK(scriptData.CovenantKey),
	}
}

func ParseBtcTx(txBytes []byte) (*wire.MsgTx, error) {
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(txBytes)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return nil, err
	}

	return &msgTx, nil
}

func SerializeBtcTx(tx *wire.MsgTx) ([]byte, error) {
	var txBuf bytes.Buffer
	if err := tx.Serialize(&txBuf); err != nil {
		return nil, err
	}
	return txBuf.Bytes(), nil
}

func ParseBtcTxFromHex(txHex string) (*wire.MsgTx, []byte, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := ParseBtcTx(txBytes)

	if err != nil {
		return nil, nil, err
	}

	return parsed, txBytes, nil
}

func GetOutputIdx(tx *wire.MsgTx, output *wire.TxOut) (uint32, error) {
	for i, txOut := range tx.TxOut {
		if bytes.Equal(txOut.PkScript, output.PkScript) && txOut.Value == output.Value {
			return uint32(i), nil
		}
	}

	return 0, fmt.Errorf("output not found")
}

func (del *BTCDelegation) GetStakingTime() uint16 {
	diff := del.EndHeight - del.StartHeight

	if diff > math.MaxUint16 {
		// In valid delegation, EndHeight is always greater than StartHeight and it is always uint16 value
		panic("invalid delegation in database")
	}

	return uint16(diff)
}
