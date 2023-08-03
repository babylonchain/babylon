package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

type BTCSlashingTx []byte

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
	sig := bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
	return &sig, nil
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

// ToMsgTxWithWitness generates a BTC slashing tx with witness from
// - the staking tx
// - validator signature
// - delegator signature
// - jury signature
func (tx *BTCSlashingTx) ToMsgTxWithWitness(stakingTx *StakingTx, valSig, delSig, jurySig *bbn.BIP340Signature) (*wire.MsgTx, error) {
	// get staking script
	stakingScript := stakingTx.StakingScript

	// get Schnorr signatures
	valSchnorrSig, err := valSig.ToBTCSig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert BTC validator signature to Schnorr signature format: %w", err)
	}
	delSchnorrSig, err := delSig.ToBTCSig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert BTC delegator signature to Schnorr signature format: %w", err)
	}
	jurySchnorrSig, err := jurySig.ToBTCSig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert jury signature to Schnorr signature format: %w", err)
	}

	// build witness from each signature
	valWitness, err := btcstaking.NewWitnessFromStakingScriptAndSignature(stakingScript, valSchnorrSig)
	if err != nil {
		return nil, fmt.Errorf("failed to build witness for BTC validator: %w", err)
	}
	delWitness, err := btcstaking.NewWitnessFromStakingScriptAndSignature(stakingScript, delSchnorrSig)
	if err != nil {
		return nil, fmt.Errorf("failed to build witness for BTC delegator: %w", err)
	}
	juryWitness, err := btcstaking.NewWitnessFromStakingScriptAndSignature(stakingScript, jurySchnorrSig)
	if err != nil {
		return nil, fmt.Errorf("failed to build witness for jury: %w", err)
	}

	// To Construct valid witness, for multisig case we need:
	// - jury signature - witnessJury[0]
	// - validator signature - witnessValidator[0]
	// - staker signature - witnessStaker[0]
	// - empty signature - which is just an empty byte array which signals we are going to use multisig.
	// 	 This must be signature on top of the stack.
	// - whole script - witnessStaker[1] (any other witness[1] will work as well)
	// - control block - witnessStaker[2] (any other witness[2] will work as well)
	slashingMsgTx, err := tx.ToMsgTx()
	if err != nil {
		return nil, fmt.Errorf("failed to convert slashing tx to Bitcoin format: %w", err)
	}
	slashingMsgTx.TxIn[0].Witness = [][]byte{
		juryWitness[0], valWitness[0], delWitness[0], []byte{}, delWitness[1], delWitness[2],
	}

	return slashingMsgTx, nil
}
