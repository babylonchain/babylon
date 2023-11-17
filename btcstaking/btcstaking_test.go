package btcstaking_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/btcstaking"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

type TestScenario struct {
	StakerKey            *btcec.PrivateKey
	ValidatorKeys        []*btcec.PrivateKey
	CovenantKeys         []*btcec.PrivateKey
	RequiredCovenantSigs uint32
	StakingAmount        btcutil.Amount
	StakingTime          uint16
}

func GenerateTestScenario(
	r *rand.Rand,
	t *testing.T,
	numValidatorKeys uint32,
	numCovenantKeys uint32,
	requiredCovenantSigs uint32,
	stakingAmount btcutil.Amount,
	stakingTime uint16,
) *TestScenario {
	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	validatorKeys := make([]*btcec.PrivateKey, numValidatorKeys)
	for i := uint32(0); i < numValidatorKeys; i++ {
		covenantPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		validatorKeys[i] = covenantPrivKey
	}

	covenantKeys := make([]*btcec.PrivateKey, numCovenantKeys)

	for i := uint32(0); i < numCovenantKeys; i++ {
		covenantPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		covenantKeys[i] = covenantPrivKey
	}

	return &TestScenario{
		StakerKey:            stakerPrivKey,
		ValidatorKeys:        validatorKeys,
		CovenantKeys:         covenantKeys,
		RequiredCovenantSigs: requiredCovenantSigs,
		StakingAmount:        stakingAmount,
		StakingTime:          stakingTime,
	}
}

func (t *TestScenario) CovenantPublicKeys() []*btcec.PublicKey {
	covenantPubKeys := make([]*btcec.PublicKey, len(t.CovenantKeys))

	for i, covenantKey := range t.CovenantKeys {
		covenantPubKeys[i] = covenantKey.PubKey()
	}

	return covenantPubKeys
}

func (t *TestScenario) ValidatorPublicKeys() []*btcec.PublicKey {
	validatorPubKeys := make([]*btcec.PublicKey, len(t.ValidatorKeys))

	for i, validatorKey := range t.ValidatorKeys {
		validatorPubKeys[i] = validatorKey.PubKey()
	}

	return validatorPubKeys
}

func TestSpendingTimeLockPath(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	scenario := GenerateTestScenario(
		r,
		t,
		1,
		5,
		3,
		btcutil.Amount(2*10e8),
		5,
	)

	stakingInfo, err := btcstaking.BuildStakingInfo(
		scenario.StakerKey.PubKey(),
		scenario.ValidatorPublicKeys(),
		scenario.CovenantPublicKeys(),
		scenario.RequiredCovenantSigs,
		scenario.StakingTime,
		scenario.StakingAmount,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(scenario.StakingAmount.MulF64(0.5)),
		},
	)

	// to spend tx as staker, we need to set the sequence number to be >= stakingTimeBlocks
	spendStakeTx.TxIn[0].Sequence = uint32(scenario.StakingTime)

	si, err := stakingInfo.TimeLockPathSpendInfo()
	require.NoError(t, err)

	sig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)

	require.NoError(t, err)

	witness, err := btcstaking.CreateBabylonWitness(
		[][]byte{sig.Serialize()},
		si,
	)

	require.NoError(t, err)

	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingInfo.StakingOutput.PkScript, stakingInfo.StakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.StakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

// generate list of signatures in valid order
func GenerateSignatures(
	t *testing.T,
	keys []*btcec.PrivateKey,
	tx *wire.MsgTx,
	stakingOutput *wire.TxOut,
	leaf txscript.TapLeaf,
) [][]byte {

	var si []*btcstaking.SignatureInfo

	for _, key := range keys {
		pubKey := key.PubKey()
		sig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
			tx,
			stakingOutput,
			key,
			leaf,
		)
		require.NoError(t, err)
		info := btcstaking.NewSignatureInfo(
			pubKey,
			sig.Serialize(),
		)
		si = append(si, info)
	}

	// sort signatures by public key
	sortedSigInfo := btcstaking.SortSignatureInfo(si)

	var sigs [][]byte = make([][]byte, len(sortedSigInfo))

	for i, sigInfo := range sortedSigInfo {
		sig := sigInfo
		sigs[i] = sig.Signature
	}

	return sigs
}

func TestSpendingUnbondingPathCovenant35MultiSig(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// we are having here 3/5 covenant threshold sig
	scenario := GenerateTestScenario(
		r,
		t,
		1,
		5,
		3,
		btcutil.Amount(2*10e8),
		5,
	)

	stakingInfo, err := btcstaking.BuildStakingInfo(
		scenario.StakerKey.PubKey(),
		scenario.ValidatorPublicKeys(),
		scenario.CovenantPublicKeys(),
		scenario.RequiredCovenantSigs,
		scenario.StakingTime,
		scenario.StakingAmount,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(scenario.StakingAmount.MulF64(0.5)),
		},
	)

	si, err := stakingInfo.UnbondingPathSpendInfo()
	require.NoError(t, err)

	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)

	require.NoError(t, err)

	// scenario where all keys are available
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	var witnessSignatures [][]byte
	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err := btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingInfo.StakingOutput.PkScript, stakingInfo.StakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.StakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)

	numOfCovenantMembers := len(scenario.CovenantKeys)
	// with each loop iteration we remove one key from the list of signatures
	for i := 0; i < numOfCovenantMembers; i++ {
		// reset signatures
		witnessSignatures = [][]byte{}

		numOfRemovedSignatures := i + 1

		covenantSigantures := GenerateSignatures(
			t,
			scenario.CovenantKeys,
			spendStakeTx,
			stakingInfo.StakingOutput,
			si.RevealedLeaf,
		)

		for j := 0; j <= i; j++ {
			// NOTE: Number provides signatures must match number of public keys in the script,
			// if we are missing some signatures those must be set to empty signature in witness
			covenantSigantures[j] = []byte{}
		}

		witnessSignatures = append(witnessSignatures, covenantSigantures...)
		witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
		witness, err := btcstaking.CreateBabylonWitness(
			witnessSignatures,
			si,
		)
		require.NoError(t, err)
		spendStakeTx.TxIn[0].Witness = witness

		if numOfCovenantMembers-numOfRemovedSignatures >= int(scenario.RequiredCovenantSigs) {
			// if we are above threshold execution should be successful
			btctest.AssertEngineExecution(t, 0, true, newEngine)
		} else {
			// we are below threshold execution should be unsuccessful
			btctest.AssertEngineExecution(t, 0, false, newEngine)
		}
	}
}

func TestSpendingUnbondingPathSingleKeyCovenant(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// generate single key covenant
	scenario := GenerateTestScenario(
		r,
		t,
		1,
		1,
		1,
		btcutil.Amount(2*10e8),
		5,
	)

	stakingInfo, err := btcstaking.BuildStakingInfo(
		scenario.StakerKey.PubKey(),
		scenario.ValidatorPublicKeys(),
		scenario.CovenantPublicKeys(),
		scenario.RequiredCovenantSigs,
		scenario.StakingTime,
		scenario.StakingAmount,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(scenario.StakingAmount.MulF64(0.5)),
		},
	)

	si, err := stakingInfo.UnbondingPathSpendInfo()
	require.NoError(t, err)

	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	// scenario where all keys are available
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	var witnessSignatures [][]byte
	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err := btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingInfo.StakingOutput.PkScript, stakingInfo.StakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.StakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

func TestSpendingSlashingPathCovenant35MultiSig(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// we are having here 3/5 covenant threshold sig
	scenario := GenerateTestScenario(
		r,
		t,
		1,
		5,
		3,
		btcutil.Amount(2*10e8),
		5,
	)

	stakingInfo, err := btcstaking.BuildStakingInfo(
		scenario.StakerKey.PubKey(),
		scenario.ValidatorPublicKeys(),
		scenario.CovenantPublicKeys(),
		scenario.RequiredCovenantSigs,
		scenario.StakingTime,
		scenario.StakingAmount,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(scenario.StakingAmount.MulF64(0.5)),
		},
	)

	si, err := stakingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	// Case without validator signature
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	var witnessSignatures [][]byte
	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err := btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingInfo.StakingOutput.PkScript, stakingInfo.StakingOutput.Value,
	)
	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.StakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	// we expect it will fail because we are missing validator signature
	btctest.AssertEngineExecution(t, 0, false, newEngine)

	// Retry with the same values but now with validator signature present
	witnessSignatures = [][]byte{}
	validatorSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.ValidatorKeys[0],
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, validatorSig.Serialize())
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err = btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	// now as we have validator signature execution should succeed
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

func TestSpendingSlashingPathCovenant35MultiSigValidatorRestaking(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// we are having here 3/5 covenant threshold sig, and we are restaking to 2 validators
	scenario := GenerateTestScenario(
		r,
		t,
		2,
		5,
		3,
		btcutil.Amount(2*10e8),
		5,
	)

	stakingInfo, err := btcstaking.BuildStakingInfo(
		scenario.StakerKey.PubKey(),
		scenario.ValidatorPublicKeys(),
		scenario.CovenantPublicKeys(),
		scenario.RequiredCovenantSigs,
		scenario.StakingTime,
		scenario.StakingAmount,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(scenario.StakingAmount.MulF64(0.5)),
		},
	)

	si, err := stakingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	// Case without validator signature
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	var witnessSignatures [][]byte
	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err := btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingInfo.StakingOutput.PkScript, stakingInfo.StakingOutput.Value,
	)
	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.StakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	// we expect it will fail because we are missing validators signature
	btctest.AssertEngineExecution(t, 0, false, newEngine)

	// Retry with the same values but now with validator signature present
	witnessSignatures = [][]byte{}

	validatorsSignatures := GenerateSignatures(
		t,
		scenario.ValidatorKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)

	// make one signature empty, script still should be valid as we require only one validator signature
	// to be present
	validatorsSignatures[0] = []byte{}

	witnessSignatures = append(witnessSignatures, covenantSigantures...)
	witnessSignatures = append(witnessSignatures, validatorsSignatures...)
	witnessSignatures = append(witnessSignatures, stakerSig.Serialize())
	witness, err = btcstaking.CreateBabylonWitness(
		witnessSignatures,
		si,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	// now as we have validator signature execution should succeed
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}
