package datagen

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func GenRandomBTCValidator(r *rand.Rand) (*bstypes.BTCValidator, error) {
	// key pairs
	btcSK, _, err := GenRandomBTCKeyPair(r)
	if err != nil {
		return nil, err
	}
	return GenRandomBTCValidatorWithBTCSK(r, btcSK)
}

func GenRandomBTCValidatorWithBTCSK(r *rand.Rand, btcSK *btcec.PrivateKey) (*bstypes.BTCValidator, error) {
	// key pairs
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	bbnSK, bbnPK, err := GenRandomSecp256k1KeyPair(r)
	if err != nil {
		return nil, err
	}
	secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
	if !ok {
		return nil, fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
	}
	// pop
	pop, err := bstypes.NewPoP(bbnSK, btcSK)
	if err != nil {
		return nil, err
	}
	return &bstypes.BTCValidator{
		BabylonPk: secp256k1PK,
		BtcPk:     bip340PK,
		Pop:       pop,
	}, nil
}

func GenRandomBTCDelegation(r *rand.Rand, valBTCPK *bbn.BIP340PubKey, startHeight uint64, endHeight uint64, totalSat uint64) (*bstypes.BTCDelegation, error) {
	// key pairs
	btcSK, btcPK, err := GenRandomBTCKeyPair(r)
	if err != nil {
		return nil, err
	}
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	bbnSK, bbnPK, err := GenRandomSecp256k1KeyPair(r)
	if err != nil {
		return nil, err
	}
	secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
	if !ok {
		return nil, fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
	}
	// pop
	pop, err := bstypes.NewPoP(bbnSK, btcSK)
	if err != nil {
		return nil, err
	}
	// TODO: generate legitimate jury signature and staking/slashing tx
	jurySchnorrSig, err := schnorr.Sign(btcSK, GenRandomByteArray(r, 32))
	if err != nil {
		return nil, err
	}
	jurySig := bbn.NewBIP340SignatureFromBTCSig(jurySchnorrSig)
	return &bstypes.BTCDelegation{
		BabylonPk:   secp256k1PK,
		BtcPk:       bip340PK,
		Pop:         pop,
		ValBtcPk:    valBTCPK,
		StartHeight: startHeight,
		EndHeight:   endHeight,
		TotalSat:    totalSat,
		JurySig:     &jurySig,
	}, nil
}

func GenBTCStakingSlashingTx(r *rand.Rand, stakerSK *btcec.PrivateKey, validatorPK, juryPK *btcec.PublicKey, stakingTimeBlocks uint16, stakingValue int64, slashingAddress string) (*bstypes.StakingTx, *bstypes.BTCSlashingTx, error) {
	btcNet := &chaincfg.SimNetParams

	stakingOutput, stakingScript, err := btcstaking.BuildStakingOutput(
		stakerSK.PubKey(),
		validatorPK,
		juryPK,
		stakingTimeBlocks,
		btcutil.Amount(stakingValue),
		btcNet,
	)
	if err != nil {
		return nil, nil, err
	}

	tx := wire.NewMsgTx(2)
	// an arbitrary input
	spend := makeSpendableOutWithRandOutPoint(r, btcutil.Amount(stakingValue+1000))
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: spend.prevOut,
		Sequence:         wire.MaxTxInSequenceNum,
		SignatureScript:  nil,
	})
	// 2 outputs for changes and staking output
	tx.AddTxOut(wire.NewTxOut(100, []byte{1, 2, 3})) // output for change, doesn't matter
	tx.AddTxOut(stakingOutput)

	// construct staking tx
	var buf bytes.Buffer
	err = tx.Serialize(&buf)
	if err != nil {
		return nil, nil, err
	}
	stakingTx := &bstypes.StakingTx{
		Tx:            buf.Bytes(),
		StakingScript: stakingScript,
	}

	// construct slashing tx
	slashingAddrBtc, err := btcutil.DecodeAddress(slashingAddress, btcNet)
	if err != nil {
		return nil, nil, err
	}
	slashingMsgTx, err := btcstaking.BuildSlashingTxFromStakingTxStrict(tx, 1, slashingAddrBtc, 10000, stakingScript, btcNet)
	if err != nil {
		return nil, nil, err
	}
	slashingTx, err := bstypes.NewBTCSlashingTxFromMsgTx(slashingMsgTx)
	if err != nil {
		return nil, nil, err
	}

	return stakingTx, slashingTx, nil
}
