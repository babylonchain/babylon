package btcstaking_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

// StakingScriptData is a struct that holds data parsed from staking script
type StakingScriptData struct {
	StakerKey    *btcec.PublicKey
	ValidatorKey *btcec.PublicKey
	CovenantKey  *btcec.PublicKey
	StakingTime  uint16
}

func NewStakingScriptData(
	stakerKey,
	validatorKey,
	covenantKey *btcec.PublicKey,
	stakingTime uint16) (*StakingScriptData, error) {

	if stakerKey == nil || validatorKey == nil || covenantKey == nil {
		return nil, fmt.Errorf("staker, validator and covenant keys cannot be nil")
	}

	return &StakingScriptData{
		StakerKey:    stakerKey,
		ValidatorKey: validatorKey,
		CovenantKey:  covenantKey,
		StakingTime:  stakingTime,
	}, nil
}

func genValidStakingScriptData(_ *testing.T, r *rand.Rand) *StakingScriptData {
	stakerPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	validatorPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	covenantPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
	stakingTime := uint16(r.Intn(math.MaxUint16))

	_, stakerPublicKey := btcec.PrivKeyFromBytes(stakerPrivKeyBytes)
	_, validatorPublicKey := btcec.PrivKeyFromBytes(validatorPrivKeyBytes)
	_, covenantPublicKey := btcec.PrivKeyFromBytes(covenantPrivKeyBytes)

	sd, _ := NewStakingScriptData(stakerPublicKey, validatorPublicKey, covenantPublicKey, stakingTime)

	return sd
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

		minStakingValue := 5000
		minFee := 2000
		// generate a random slashing rate with random precision,
		// this will include both valid and invalid ranges, so we can test both cases
		randomPrecision := r.Int63n(4)                                                                   // [0,3]
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 1001)), randomPrecision) // [0,1000] / 10^{randomPrecision}

		for i := 0; i < stakingTxNumOutputs; i++ {
			if i == stakingOutputIdx {
				info, err := btcstaking.BuildStakingInfo(
					sd.StakerKey,
					[]*btcec.PublicKey{sd.ValidatorKey},
					[]*btcec.PublicKey{sd.CovenantKey},
					1,
					sd.StakingTime,
					btcutil.Amount(r.Intn(5000)+minStakingValue),
					&chaincfg.MainNetParams,
				)

				require.NoError(t, err)
				stakingTx.AddTxOut(info.StakingOutput)
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
		testSlashingTx(r, t, stakingTx, stakingOutputIdx, slashingRate, int64(minFee))

		// Check case with some random fee
		fee := int64(r.Intn(1000) + minFee)
		testSlashingTx(r, t, stakingTx, stakingOutputIdx, slashingRate, fee)

	})
}

func genRandomBTCAddress(r *rand.Rand) (*btcutil.AddressPubKeyHash, error) {
	return btcutil.NewAddressPubKeyHash(datagen.GenRandomByteArray(r, 20), &chaincfg.MainNetParams)
}

func testSlashingTx(r *rand.Rand, t *testing.T, stakingTx *wire.MsgTx, stakingOutputIdx int, slashingRate sdkmath.LegacyDec, fee int64) {
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
			err = btcstaking.CheckTransactions(
				slashingTx,
				stakingTx,
				uint32(stakingOutputIdx),
				fee,
				slashingRate,
				slashingAddress,
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
