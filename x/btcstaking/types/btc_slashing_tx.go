package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/babylonchain/babylon/btcstaking"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
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
	return bbn.NewBTCTxFromBytes(*tx)
}

func (tx *BTCSlashingTx) MustGetTxHash() *chainhash.Hash {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		panic(err)
	}
	txHash := msgTx.TxHash()
	return &txHash
}

// Sign generates a signature on the slashing tx
func (tx *BTCSlashingTx) Sign(
	fundingTx *wire.MsgTx,
	spendOutputIndex uint32,
	slashingPkScriptPath []byte,
	sk *btcec.PrivateKey,
) (*bbn.BIP340Signature, error) {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	schnorrSig, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
		msgTx,
		fundingTx,
		spendOutputIndex,
		slashingPkScriptPath,
		sk,
	)
	if err != nil {
		return nil, err
	}
	return bbn.NewBIP340SignatureFromBTCSig(schnorrSig), nil
}

// VerifySignature verifies a signature on the slashing tx signed by staker, finality provider, or covenant
func (tx *BTCSlashingTx) VerifySignature(
	fundingPkScript []byte,
	fundingAmount int64,
	slashingPkScriptPath []byte,
	pk *btcec.PublicKey,
	sig *bbn.BIP340Signature,
) error {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return err
	}
	return btcstaking.VerifyTransactionSigWithOutputData(
		msgTx,
		fundingPkScript,
		fundingAmount,
		slashingPkScriptPath,
		pk,
		*sig,
	)
}

// EncSign generates an adaptor signature on the slashing tx with finality provider's
// public key as encryption key
func (tx *BTCSlashingTx) EncSign(
	fundingMsgTx *wire.MsgTx,
	spendOutputIndex uint32,
	slashingPkScriptPath []byte,
	sk *btcec.PrivateKey,
	encKey *asig.EncryptionKey,
) (*asig.AdaptorSignature, error) {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	adaptorSig, err := btcstaking.EncSignTxWithOneScriptSpendInputStrict(
		msgTx,
		fundingMsgTx,
		spendOutputIndex,
		slashingPkScriptPath,
		sk,
		encKey,
	)
	if err != nil {
		return nil, err
	}

	return adaptorSig, nil
}

// EncVerifyAdaptorSignature verifies an adaptor signature on the slashing tx
// with the finality provider's public key as encryption key
func (tx *BTCSlashingTx) EncVerifyAdaptorSignature(
	stakingPkScript []byte,
	stakingAmount int64,
	slashingPkScriptPath []byte,
	pk *btcec.PublicKey,
	encKey *asig.EncryptionKey,
	sig *asig.AdaptorSignature,
) error {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return err
	}
	return btcstaking.EncVerifyTransactionSigWithOutputData(
		msgTx,
		stakingPkScript,
		stakingAmount,
		slashingPkScriptPath,
		pk,
		encKey,
		sig,
	)
}

// ParseEncVerifyAdaptorSignatures verifies a list of adaptor signatures, each
// encrypted by a restaked validator PK and signed by the given PK, w.r.t. the
// given funding output (in staking or unbonding tx), slashing spend info and
// slashing tx
// It returns a list of parsed adaptor signatures in case of successful verification
func (tx *BTCSlashingTx) ParseEncVerifyAdaptorSignatures(
	fundingOut *wire.TxOut,
	slashingSpendInfo *btcstaking.SpendInfo,
	pk *bbn.BIP340PubKey,
	valPKs []bbn.BIP340PubKey,
	sigs [][]byte,
) ([]asig.AdaptorSignature, error) {
	var adaptorSigs []asig.AdaptorSignature = make([]asig.AdaptorSignature, len(sigs))
	for i := range sigs {
		sig := sigs[i]
		adaptorSig, err := asig.NewAdaptorSignatureFromBytes(sig)
		if err != nil {
			return nil, err
		}
		encKey, err := asig.NewEncryptionKeyFromBTCPK(valPKs[i].MustToBTCPK())
		if err != nil {
			return nil, err
		}
		err = tx.EncVerifyAdaptorSignature(
			fundingOut.PkScript,
			fundingOut.Value,
			slashingSpendInfo.GetPkScriptPath(),
			pk.MustToBTCPK(),
			encKey,
			adaptorSig,
		)
		if err != nil {
			return nil, ErrInvalidCovenantSig.Wrapf("err: %v", err)
		}
		adaptorSigs[i] = *adaptorSig
	}
	return adaptorSigs, nil
}

// EncVerifyAdaptorSignatures verifies a list of adaptor signatures, each
// encrypted by a restaked validator PK and signed by the given PK, w.r.t. the
// given funding output (in staking or unbonding tx), slashing spend info and
// slashing tx
func (tx *BTCSlashingTx) EncVerifyAdaptorSignatures(
	fundingOut *wire.TxOut,
	slashingSpendInfo *btcstaking.SpendInfo,
	pk *bbn.BIP340PubKey,
	valPKs []bbn.BIP340PubKey,
	sigs [][]byte,
) error {

	_, err := tx.ParseEncVerifyAdaptorSignatures(fundingOut, slashingSpendInfo, pk, valPKs, sigs)
	if err != nil {
		return err
	}

	return nil

}

func (tx *BTCSlashingTx) BuildSlashingTxWithWitness(
	fpSK *btcec.PrivateKey,
	fundingMsgTx *wire.MsgTx,
	outputIdx uint32,
	delegatorSig *bbn.BIP340Signature,
	covenantSigs []*asig.AdaptorSignature,
	slashingPathSpendInfo *btcstaking.SpendInfo,
) (*wire.MsgTx, error) {
	fpSig, err := tx.Sign(fundingMsgTx, outputIdx, slashingPathSpendInfo.GetPkScriptPath(), fpSK)
	if err != nil {
		return nil, fmt.Errorf("failed to sign slashing tx for the finality provider: %w", err)
	}

	// decrypt covenant adaptor signature to Schnorr signature using finality provider's SK,
	// then marshal
	decKey, err := asig.NewDecyptionKeyFromBTCSK(fpSK)
	if err != nil {
		return nil, fmt.Errorf("failed to get decryption key from BTC SK: %w", err)
	}

	var covSigs []*schnorr.Signature
	for _, covenantSig := range covenantSigs {
		if covenantSig != nil {
			covSigs = append(covSigs, covenantSig.Decrypt(decKey))
		} else {
			covSigs = append(covSigs, nil)
		}
	}

	// construct witness
	witness, err := slashingPathSpendInfo.CreateSlashingPathWitness(
		covSigs,
		[]*schnorr.Signature{fpSig.MustToBTCSig()}, // TODO: work with restaking
		delegatorSig.MustToBTCSig(),
	)
	if err != nil {
		return nil, err
	}
	// add witness to slashing tx
	slashingMsgTxWithWitness, err := tx.ToMsgTx()
	if err != nil {
		return nil, err
	}
	slashingMsgTxWithWitness.TxIn[0].Witness = witness

	return slashingMsgTxWithWitness, nil
}
