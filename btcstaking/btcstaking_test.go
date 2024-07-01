package btcstaking_test

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/babylonchain/babylon/btcstaking"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

type TestScenario struct {
	StakerKey            *btcec.PrivateKey
	FinalityProviderKeys []*btcec.PrivateKey
	CovenantKeys         []*btcec.PrivateKey
	RequiredCovenantSigs uint32
	StakingAmount        btcutil.Amount
	StakingTime          uint16
}

func GenerateTestScenario(
	r *rand.Rand,
	t *testing.T,
	numFinalityProviderKeys uint32,
	numCovenantKeys uint32,
	requiredCovenantSigs uint32,
	stakingAmount btcutil.Amount,
	stakingTime uint16,
) *TestScenario {
	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	finalityProviderKeys := make([]*btcec.PrivateKey, numFinalityProviderKeys)
	for i := uint32(0); i < numFinalityProviderKeys; i++ {
		finalityProviderPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		finalityProviderKeys[i] = finalityProviderPrivKey
	}

	covenantKeys := make([]*btcec.PrivateKey, numCovenantKeys)

	for i := uint32(0); i < numCovenantKeys; i++ {
		covenantPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		covenantKeys[i] = covenantPrivKey
	}

	return &TestScenario{
		StakerKey:            stakerPrivKey,
		FinalityProviderKeys: finalityProviderKeys,
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

func (t *TestScenario) FinalityProviderPublicKeys() []*btcec.PublicKey {
	finalityProviderPubKeys := make([]*btcec.PublicKey, len(t.FinalityProviderKeys))

	for i, fpKey := range t.FinalityProviderKeys {
		finalityProviderPubKeys[i] = fpKey.PubKey()
	}

	return finalityProviderPubKeys
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
		scenario.FinalityProviderPublicKeys(),
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

	witness, err := si.CreateTimeLockPathWitness(sig)
	require.NoError(t, err)

	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := stakingInfo.GetOutputFetcher()

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.GetPkScript(),
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

type SignatureInfo struct {
	SignerPubKey *btcec.PublicKey
	Signature    *schnorr.Signature
}

func NewSignatureInfo(
	signerPubKey *btcec.PublicKey,
	signature *schnorr.Signature,
) *SignatureInfo {
	return &SignatureInfo{
		SignerPubKey: signerPubKey,
		Signature:    signature,
	}
}

// Helper function to sort all signatures in reverse lexicographical order of signing public keys
// this way signatures are ready to be used in multisig witness with corresponding public keys
func sortSignatureInfo(infos []*SignatureInfo) []*SignatureInfo {
	sortedInfos := make([]*SignatureInfo, len(infos))
	copy(sortedInfos, infos)
	sort.SliceStable(sortedInfos, func(i, j int) bool {
		keyIBytes := schnorr.SerializePubKey(sortedInfos[i].SignerPubKey)
		keyJBytes := schnorr.SerializePubKey(sortedInfos[j].SignerPubKey)
		return bytes.Compare(keyIBytes, keyJBytes) == 1
	})

	return sortedInfos
}

// generate list of signatures in valid order
func GenerateSignatures(
	t *testing.T,
	keys []*btcec.PrivateKey,
	tx *wire.MsgTx,
	stakingOutput *wire.TxOut,
	leaf txscript.TapLeaf,
) []*schnorr.Signature {

	var si []*SignatureInfo

	for _, key := range keys {
		pubKey := key.PubKey()
		sig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
			tx,
			stakingOutput,
			key,
			leaf,
		)
		require.NoError(t, err)
		info := NewSignatureInfo(
			pubKey,
			sig,
		)
		si = append(si, info)
	}

	// sort signatures by public key
	sortedSigInfo := sortSignatureInfo(si)

	var sigs []*schnorr.Signature = make([]*schnorr.Signature, len(sortedSigInfo))

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
		scenario.FinalityProviderPublicKeys(),
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

	covenantSigantures[1] = nil
	covenantSigantures[3] = nil

	witness, err := si.CreateUnbondingPathWitness(covenantSigantures, stakerSig)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := stakingInfo.GetOutputFetcher()

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.GetPkScript(),
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)

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
		scenario.FinalityProviderPublicKeys(),
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
	witness, err := si.CreateUnbondingPathWitness(covenantSigantures, stakerSig)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := stakingInfo.GetOutputFetcher()

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.GetPkScript(),
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
		scenario.FinalityProviderPublicKeys(),
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

	// generate staker signature, covenant signatures, and finality provider signature
	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)
	require.NoError(t, err)
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	fpSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.FinalityProviderKeys[0],
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	covenantSigantures[0] = nil
	covenantSigantures[3] = nil

	witness, err := si.CreateSlashingPathWitness(
		covenantSigantures,
		[]*schnorr.Signature{fpSig},
		stakerSig,
	)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	// now as we have finality provider signature execution should succeed
	prevOutputFetcher := stakingInfo.GetOutputFetcher()
	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.GetPkScript(),
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

func TestSpendingSlashingPathCovenant35MultiSigFinalityProviderRestaking(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// we have 3 out of 5 covenant committee, and we are restaking to 2 finality providers
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
		scenario.FinalityProviderPublicKeys(),
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

	// generate staker signature, covenant signatures, and finality provider signature
	stakerSig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		stakingInfo.StakingOutput,
		scenario.StakerKey,
		si.RevealedLeaf,
	)
	require.NoError(t, err)

	// only use 3 out of 5 covenant signatures
	covenantSigantures := GenerateSignatures(
		t,
		scenario.CovenantKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	covenantSigantures[0] = nil
	covenantSigantures[1] = nil

	// only use one of the finality provider signatures
	// script should still be valid as we require only one finality provider signature
	// to be present
	fpSignatures := GenerateSignatures(
		t,
		scenario.FinalityProviderKeys,
		spendStakeTx,
		stakingInfo.StakingOutput,
		si.RevealedLeaf,
	)
	fpSignatures[0] = nil

	witness, err := si.CreateSlashingPathWitness(covenantSigantures, fpSignatures, stakerSig)
	require.NoError(t, err)
	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := stakingInfo.GetOutputFetcher()
	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingInfo.GetPkScript(),
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingInfo.StakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

func TestSpendingRelativeTimeLockScript(t *testing.T) {
	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	stakerPubKey := stakerPrivKey.PubKey()
	lockTime := uint16(10)
	lockedAmount := btcutil.Amount(2 * 10e8)

	// to spend output with relative timelock transaction need to be version two or higher
	spendStakeTx := wire.NewMsgTx(2)
	spendStakeTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			// spend half of the staking amount
			Value: int64(lockedAmount.MulF64(0.5)),
		},
	)

	tls, err := btcstaking.BuildRelativeTimelockTaprootScript(
		stakerPubKey,
		lockTime,
		&chaincfg.MainNetParams,
	)
	require.NoError(t, err)

	timeLockOutput := &wire.TxOut{
		PkScript: tls.PkScript,
		Value:    int64(lockedAmount),
	}

	// we need to set sequence number before signing, as signing commits to sequence
	// number
	spendStakeTx.TxIn[0].Sequence = uint32(tls.LockTime)

	sig, err := btcstaking.SignTxWithOneScriptSpendInputFromTapLeaf(
		spendStakeTx,
		timeLockOutput,
		stakerPrivKey,
		tls.SpendInfo.RevealedLeaf,
	)

	require.NoError(t, err)

	witness, err := btcstaking.CreateWitness(
		tls.SpendInfo,
		[][]byte{sig.Serialize()},
	)

	require.NoError(t, err)

	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		timeLockOutput.PkScript, timeLockOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			timeLockOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), timeLockOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}
