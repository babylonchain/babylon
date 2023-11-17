package btcstaking_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/btcstaking"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func genValidStakingScriptData(t *testing.T, r *rand.Rand) *btcstaking.StakingScriptData {
	stakerPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	validatorPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	covenantPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	stakingTime := uint16(r.Intn(math.MaxUint16))

	_, stakerPublicKey := btcec.PrivKeyFromBytes(stakerPrivKeyBytes)
	_, validatorPublicKey := btcec.PrivKeyFromBytes(validatorPrivKeyBytes)
	_, covenantPublicKey := btcec.PrivKeyFromBytes(covenantPrivKeyBytes)

	sd, _ := btcstaking.NewStakingScriptData(stakerPublicKey, validatorPublicKey, covenantPublicKey, stakingTime)

	return sd
}

func FuzzGeneratingParsingValidStakingScript(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sd := genValidStakingScriptData(t, r)

		script, err := sd.BuildStakingScript()
		require.NoError(t, err)
		parsedScript, err := btcstaking.ParseStakingTransactionScript(script)
		require.NoError(t, err)

		require.Equal(t, parsedScript.StakingTime, sd.StakingTime)
		require.Equal(t, schnorr.SerializePubKey(sd.StakerKey), schnorr.SerializePubKey(parsedScript.StakerKey))
		require.Equal(t, schnorr.SerializePubKey(sd.ValidatorKey), schnorr.SerializePubKey(parsedScript.ValidatorKey))
		require.Equal(t, schnorr.SerializePubKey(sd.CovenantKey), schnorr.SerializePubKey(parsedScript.CovenantKey))
	})
}

func FuzzGeneratingValidStakingSlashingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// we do not care for inputs in staking tx
		stakingTx := wire.NewMsgTx(2)
		stakingOutputIdx := r.Intn(5)
		// always more outputs than stakingOutputIdx
		stakingTxNumOutputs := r.Intn(5) + 10
		sd := genValidStakingScriptData(t, r)
		script, err := sd.BuildStakingScript()
		require.NoError(t, err)
		minStakingValue := 5000
		minFee := 2000
		// generate a random slashing rate with random precision,
		// this will include both valid and invalid ranges, so we can test both cases
		randomPrecision := r.Int63n(4) // [0,3]
		slashingRate := sdk.NewDecWithPrec(int64(datagen.RandomInt(r, 1001)), randomPrecision) // [0,1000] / 10^{randomPrecision}

		for i := 0; i < stakingTxNumOutputs; i++ {
			if i == stakingOutputIdx {
				stakingOutput, _, err := btcstaking.BuildStakingOutput(
					sd.StakerKey,
					sd.ValidatorKey,
					sd.CovenantKey,
					sd.StakingTime,
					btcutil.Amount(r.Intn(5000)+minStakingValue),
					&chaincfg.MainNetParams,
				)
				require.NoError(t, err)
				stakingTx.AddTxOut(stakingOutput)
			} else {
				stakingTx.AddTxOut(
					&wire.TxOut{
						PkScript: datagen.GenRandomByteArray(r, 32),
						Value:    int64(r.Intn(5000) + 1),
					},
				)
			}
		}

		// Always check case with min fee
		testSlashingTx(r, t, stakingTx, stakingOutputIdx, slashingRate, script, int64(minFee))

		// Check case with some random fee
		fee := int64(r.Intn(1000) + minFee)
		testSlashingTx(r, t, stakingTx, stakingOutputIdx, slashingRate, script, fee)

	})
}

func genRandomBTCAddress(r *rand.Rand) (*btcutil.AddressPubKeyHash, error) {
	return btcutil.NewAddressPubKeyHash(datagen.GenRandomByteArray(r, 20), &chaincfg.MainNetParams)
}

func testSlashingTx(r *rand.Rand, t *testing.T, stakingTx *wire.MsgTx, stakingOutputIdx int, slashingRate sdk.Dec,
	script []byte, fee int64) {
	dustThreshold := 546 // in satoshis

	// Generate random slashing and change addresses
	slashingAddress, err := genRandomBTCAddress(r)
	require.NoError(t, err)

	changeAddress, err := genRandomBTCAddress(r)
	require.NoError(t, err)

	// Construct slashing transaction using the provided parameters
	slashingTx, err := btcstaking.BuildSlashingTxFromStakingTxStrict(
		stakingTx,
		uint32(stakingOutputIdx),
		slashingAddress, changeAddress,
		fee,
		slashingRate,
		script,
		&chaincfg.MainNetParams,
	)

	if btcstaking.IsSlashingRateValid(slashingRate) {
		// If the slashing rate is valid i.e., in the range (0,1) with at most 2 decimal places,
		// it is still possible that the slashing transaction is invalid. The following checks will confirm that
		// slashing tx is not constructed if
		// - the change output has insufficient funds.
		// - the change output is less than the dust threshold.
		// - The slashing output is less than the dust threshold.

		slashingRateFloat64, err2 := slashingRate.Float64()
		require.NoError(t, err2)

		stakingAmount := btcutil.Amount(stakingTx.TxOut[stakingOutputIdx].Value)
		slashingAmount := stakingAmount.MulF64(slashingRateFloat64)
		changeAmount := stakingAmount - slashingAmount - btcutil.Amount(fee)

		if changeAmount <= 0 {
			require.Error(t, err)
			require.ErrorIs(t, err, btcstaking.ErrInsufficientChangeAmount)
		} else if changeAmount <= btcutil.Amount(dustThreshold) || slashingAmount <= btcutil.Amount(dustThreshold) {
			require.Error(t, err)
			require.ErrorIs(t, err, btcstaking.ErrDustOutputFound)
		} else {
			require.NoError(t, err)
			_, err = btcstaking.CheckTransactions(
				slashingTx,
				stakingTx,
				fee,
				slashingRate,
				slashingAddress,
				script,
				&chaincfg.MainNetParams,
			)
			require.NoError(t, err)
		}
	} else {
		require.Error(t, err)
		require.ErrorIs(t, err, btcstaking.ErrInvalidSlashingRate)
	}
}

func FuzzGeneratingSignatureValidation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		pk, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		inputHash, err := chainhash.NewHash(datagen.GenRandomByteArray(r, 32))
		require.NoError(t, err)

		tx := wire.NewMsgTx(2)
		foundingOutput := wire.NewTxOut(int64(r.Intn(1000)), datagen.GenRandomByteArray(r, 32))
		tx.AddTxIn(
			wire.NewTxIn(wire.NewOutPoint(inputHash, uint32(r.Intn(20))), nil, nil),
		)
		tx.AddTxOut(
			wire.NewTxOut(int64(r.Intn(1000)), datagen.GenRandomByteArray(r, 32)),
		)
		script := datagen.GenRandomByteArray(r, 150)

		sig, err := btcstaking.SignTxWithOneScriptSpendInputFromScript(
			tx,
			foundingOutput,
			pk,
			script,
		)

		require.NoError(t, err)

		err = btcstaking.VerifyTransactionSigWithOutput(
			tx,
			foundingOutput,
			script,
			pk.PubKey(),
			sig.Serialize(),
		)

		require.NoError(t, err)
	})
}

func TestStakingScriptExecutionSingleStaker(t *testing.T) {
	const (
		stakingValue      = btcutil.Amount(2 * 10e8)
		stakingTimeBlocks = 5
	)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	validatorPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	covenantPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	txid, err := chainhash.NewHash(datagen.GenRandomByteArray(r, 32))
	require.NoError(t, err)

	stakingOut := &wire.OutPoint{
		Hash:  *txid,
		Index: 0,
	}

	stakingOutput, stakingScript, err := btcstaking.BuildStakingOutput(
		stakerPrivKey.PubKey(),
		validatorPrivKey.PubKey(),
		covenantPrivKey.PubKey(),
		stakingTimeBlocks,
		stakingValue,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)

	spendStakeTx.AddTxIn(wire.NewTxIn(stakingOut, nil, nil))

	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			Value:    1 * 10e8,
		},
	)

	// to spend tx as staker, we need to set the sequence number to be >= stakingTimeBlocks
	spendStakeTx.TxIn[0].Sequence = stakingTimeBlocks

	witness, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		stakerPrivKey,
	)

	require.NoError(t, err)

	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingOutput.PkScript, stakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}

func TestStakingScriptExecutionMulitSig(t *testing.T) {
	const (
		stakingValue      = btcutil.Amount(2 * 10e8)
		stakingTimeBlocks = 5
	)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	validatorPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	covenantPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	txid, err := chainhash.NewHash(datagen.GenRandomByteArray(r, 32))
	require.NoError(t, err)

	stakingOut := &wire.OutPoint{
		Hash:  *txid,
		Index: 0,
	}

	stakingOutput, stakingScript, err := btcstaking.BuildStakingOutput(
		stakerPrivKey.PubKey(),
		validatorPrivKey.PubKey(),
		covenantPrivKey.PubKey(),
		stakingTimeBlocks,
		stakingValue,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)

	spendStakeTx.AddTxIn(wire.NewTxIn(stakingOut, nil, nil))

	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			Value:    1 * 10e8,
		},
	)

	witnessStaker, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		stakerPrivKey,
	)
	require.NoError(t, err)

	witnessValidator, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		validatorPrivKey,
	)

	require.NoError(t, err)

	witnessCovenant, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		covenantPrivKey,
	)

	require.NoError(t, err)

	// To Construct valid witness, for multisig case we need:
	// - covenant signature - witnessCovenant[0]
	// - validator signature - witnessValidator[0]
	// - staker signature - witnessStaker[0]
	// - empty signature - which is just an empty byte array which signals we are going to use multisig.
	// 	 This must be signagure on top of the stack.
	// - whole script - witnessStaker[1] (any other wittness[1] will work as well)
	// - control block - witnessStaker[2] (any other wittness[2] will work as well)
	spendStakeTx.TxIn[0].Witness = [][]byte{
		witnessCovenant[0], witnessValidator[0], witnessStaker[0], []byte{}, witnessStaker[1], witnessStaker[2],
	}

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingOutput.PkScript, stakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingOutput.Value,
			prevOutputFetcher,
		)
	}
	btctest.AssertEngineExecution(t, 0, true, newEngine)
}
