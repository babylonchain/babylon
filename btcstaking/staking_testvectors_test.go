package btcstaking_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func getBtcNetworkParams(network string) (*chaincfg.Params, error) {
	switch network {
	case "testnet3":
		return &chaincfg.TestNet3Params, nil
	case "mainnet":
		return &chaincfg.MainNetParams, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	case "signet":
		return &chaincfg.SigNetParams, nil
	default:
		return nil, fmt.Errorf("unknown network %s", network)
	}
}

func serializeBTCTx(tx *wire.MsgTx) ([]byte, error) {
	var txBuf bytes.Buffer
	if err := tx.Serialize(&txBuf); err != nil {
		return nil, err
	}
	return txBuf.Bytes(), nil
}

func serializeBTCTxToHex(tx *wire.MsgTx) (string, error) {
	bytes, err := serializeBTCTx(tx)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func ReadTestCases() *TestCases {
	// Open the JSON file
	file, err := os.Open("./testvectors/vectors.json")
	if err != nil {
		panic(fmt.Errorf("Error opening file: %w", err))
	}
	defer file.Close()

	// Create a decoder
	decoder := json.NewDecoder(file)

	// Create a variable of the type of your struct
	var cases TestCases

	// Decode the JSON data into the struct
	if err := decoder.Decode(&cases); err != nil {
		panic(fmt.Sprintf("Error decoding JSON: %s", err))
	}

	return &cases
}

type Parameters struct {
	CovenantPublicKeys         []string `json:"covenant_public_keys"`
	CovenantQuorum             int      `json:"covenant_quorum"`
	FinalityProviderPublicKeys []string `json:"finality_provider_public_keys"`
	StakerPublicKey            string   `json:"staker_public_key"`
	StakingTime                int      `json:"staking_time"`
	StakingValue               int      `json:"staking_value"`
	StakingTxHash              string   `json:"staking_tx_hash"`
	StakingOutputIndex         int      `json:"staking_output_index"`
	UnbondingTxVersion         int      `json:"unbonding_tx_version"`
	UnbondingTime              int      `json:"unbonding_time"`
	UnbondingFee               int      `json:"unbonding_fee"`
	MagicBytes                 string   `json:"magic_bytes"`
	Network                    string   `json:"network"`
}

type Expected struct {
	StakingOutputPkScript              string `json:"staking_output_pkscript_hex"`
	StakingOutputValue                 int    `json:"staking_output_value"`
	StakingTransactionTimeLockScript   string `json:"staking_transaction_timelock_script_hex"`
	StakingTransactionUnbondingScript  string `json:"staking_transaction_unbonding_script_hex"`
	StakingTransactionSlashingScript   string `json:"staking_transaction_slashing_script_hex"`
	UnbondingTransactionHex            string `json:"unbonding_transaction_hex"`
	UnbondingTransactionTimeLockScript string `json:"unbonding_transaction_time_lock_script_hex"`
	UnbondingTransactionSlashingScript string `json:"unbonding_transaction_slashing_script_hex"`
	OpReturnScript                     string `json:"op_return_script_hex"`
}

type TestCase struct {
	Description string      `json:"name"`
	Parameters  *Parameters `json:"parameters"`
	Expected    *Expected   `json:"expected"`
}

type TestCases struct {
	Test []TestCase `json:"vectors"`
}

type ParsedParams struct {
	CovenantPublicKeys         []*btcec.PublicKey
	CovenantQuorum             uint32
	FinalityProviderPublicKeys []*btcec.PublicKey
	StakerPublicKey            *btcec.PublicKey
	StakingTime                uint16
	StakingValue               btcutil.Amount
	StakingTxHash              *chainhash.Hash
	StakingOutputIndex         uint32
	UnbondingTxVersion         uint32
	UnbondingTime              uint16
	UnbondingFee               btcutil.Amount
	MagicBytes                 []byte
	Network                    *chaincfg.Params
}

// function which parses Parameters
func parseTestParams(t *testing.T, p *Parameters) (*ParsedParams, error) {
	covenantKeys := keysToPubKeys(t, p.CovenantPublicKeys)
	finalityKeys := keysToPubKeys(t, p.FinalityProviderPublicKeys)
	stakerPk := keysToPubKeys(t, []string{p.StakerPublicKey})[0]

	stakingTxHash, err := chainhash.NewHashFromStr(p.StakingTxHash)
	if err != nil {
		return nil, err
	}

	magicBytes, err := hex.DecodeString(p.MagicBytes)
	if err != nil {
		return nil, err
	}

	network, err := getBtcNetworkParams(p.Network)
	if err != nil {
		return nil, err
	}

	return &ParsedParams{
		CovenantPublicKeys:         covenantKeys,
		CovenantQuorum:             uint32(p.CovenantQuorum),
		FinalityProviderPublicKeys: finalityKeys,
		StakerPublicKey:            stakerPk,
		StakingTime:                uint16(p.StakingTime),
		StakingValue:               btcutil.Amount(p.StakingValue),
		StakingTxHash:              stakingTxHash,
		StakingOutputIndex:         uint32(p.StakingOutputIndex),
		UnbondingTxVersion:         uint32(p.UnbondingTxVersion),
		UnbondingTime:              uint16(p.UnbondingTime),
		UnbondingFee:               btcutil.Amount(p.UnbondingFee),
		MagicBytes:                 magicBytes,
		Network:                    network,
	}, nil
}

func TestVectorsCompatiblity(t *testing.T) {
	cases := ReadTestCases()

	for _, tc := range cases.Test {
		t.Logf("Running test case: %s", tc.Description)
		parsedParams, err := parseTestParams(t, tc.Parameters)

		if err != nil {
			require.NoError(t, fmt.Errorf("error parsing test parameters for case %s: %w", tc.Description, err))
		}

		info, err := btcstaking.BuildStakingInfo(
			parsedParams.StakerPublicKey,
			parsedParams.FinalityProviderPublicKeys,
			parsedParams.CovenantPublicKeys,
			parsedParams.CovenantQuorum,
			parsedParams.StakingTime,
			parsedParams.StakingValue,
			parsedParams.Network,
		)

		if err != nil {
			require.NoError(t, fmt.Errorf("error building staking info for case %s: %w", tc.Description, err))
		}

		sti, err := info.TimeLockPathSpendInfo()

		if err != nil {
			require.NoError(t, fmt.Errorf("error building staking timelock path spend info for case %s: %w", tc.Description, err))
		}

		sui, err := info.UnbondingPathSpendInfo()

		if err != nil {
			require.NoError(t, fmt.Errorf("error building staking unbonding path spend info for case %s: %w", tc.Description, err))
		}

		ssi, err := info.SlashingPathSpendInfo()

		if err != nil {
			require.NoError(t, fmt.Errorf("error building staking slashing path spend info for case %s: %w", tc.Description, err))
		}

		ubInfo, err := btcstaking.BuildUnbondingInfo(
			parsedParams.StakerPublicKey,
			parsedParams.FinalityProviderPublicKeys,
			parsedParams.CovenantPublicKeys,
			parsedParams.CovenantQuorum,
			parsedParams.UnbondingTime,
			parsedParams.StakingValue-parsedParams.UnbondingFee,
			parsedParams.Network,
		)

		if err != nil {
			require.NoError(t, fmt.Errorf("error building unbonding info for case %s: %w", tc.Description, err))
		}

		uti, err := ubInfo.TimeLockPathSpendInfo()
		if err != nil {
			require.NoError(t, fmt.Errorf("error building unbonding timelock path spend info for case %s: %w", tc.Description, err))
		}
		usi, err := ubInfo.SlashingPathSpendInfo()

		if err != nil {
			require.NoError(t, fmt.Errorf("error building unbonding slashing path spend info for case %s: %w", tc.Description, err))
		}

		ubtTx := wire.NewMsgTx(2)
		ubtTx.AddTxIn(wire.NewTxIn(
			wire.NewOutPoint(
				parsedParams.StakingTxHash,
				parsedParams.StakingOutputIndex,
			),
			nil,
			nil,
		))
		ubtTx.AddTxOut(ubInfo.UnbondingOutput)

		serializedUbtTx, err := serializeBTCTx(ubtTx)
		if err != nil {
			require.NoError(t, fmt.Errorf("error serializing unbonding tx for case %s: %w", tc.Description, err))
		}

		require.Equal(t, tc.Expected.StakingOutputPkScript, hex.EncodeToString(info.StakingOutput.PkScript), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.StakingOutputValue, int(info.StakingOutput.Value), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.StakingTransactionTimeLockScript, hex.EncodeToString(sti.RevealedLeaf.Script), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.StakingTransactionUnbondingScript, hex.EncodeToString(sui.RevealedLeaf.Script), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.StakingTransactionSlashingScript, hex.EncodeToString(ssi.RevealedLeaf.Script), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.UnbondingTransactionHex, hex.EncodeToString(serializedUbtTx), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.UnbondingTransactionTimeLockScript, hex.EncodeToString(uti.RevealedLeaf.Script), fmt.Sprintf("failed case: %s", tc.Description))
		require.Equal(t, tc.Expected.UnbondingTransactionSlashingScript, hex.EncodeToString(usi.RevealedLeaf.Script), fmt.Sprintf("failed case: %s", tc.Description))

		if tc.Expected.OpReturnScript != "" {
			data, err := btcstaking.NewV0OpReturnDataFromParsed(
				parsedParams.MagicBytes,
				parsedParams.StakerPublicKey,
				parsedParams.FinalityProviderPublicKeys[0],
				parsedParams.StakingTime,
			)

			if err != nil {
				require.NoError(t, fmt.Errorf("error building op_return data for case %s: %w", tc.Description, err))
			}

			opReturnOutput, err := data.ToTxOutput()

			if err != nil {
				require.NoError(t, fmt.Errorf("error building op_return output for case %s: %w", tc.Description, err))
			}

			require.Equal(t, tc.Expected.OpReturnScript, hex.EncodeToString(opReturnOutput.PkScript), fmt.Sprintf("failed case: %s", tc.Description))
		}
	}
}

func generateKeys(t *testing.T, num int) []string {
	var keys []string

	for i := 0; i < num; i++ {
		k, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		keys = append(keys, hex.EncodeToString(k.PubKey().SerializeCompressed()))
	}
	return keys
}

func keysToPubKeys(t *testing.T, keys []string) []*btcec.PublicKey {
	var pks []*btcec.PublicKey

	for _, key := range keys {
		b, err := hex.DecodeString(key)
		require.NoError(t, err)

		pk, err := btcec.ParsePubKey(b)
		require.NoError(t, err)

		pks = append(pks, pk)
	}

	return pks
}

// helper to easily generate test cases
func GenerateTestCase(
	t *testing.T,
	desc string,
	numCovenantKeys int,
	covenantQuorum int,
	numFinalityKeys int,
	stakingAmout int,
	stakingTime int,
	unbondingTime int,
	unbondingFee int,
	magicBytes []byte,
) string {
	emptyHash := [32]byte{}
	eh, err := chainhash.NewHash(emptyHash[:])
	require.NoError(t, err)
	covenantKeys := generateKeys(t, numCovenantKeys)
	finalityKeys := generateKeys(t, numFinalityKeys)
	stakerKeys := generateKeys(t, 1)

	info, err := btcstaking.BuildStakingInfo(
		keysToPubKeys(t, stakerKeys)[0],
		keysToPubKeys(t, finalityKeys),
		keysToPubKeys(t, covenantKeys),
		uint32(covenantQuorum),
		uint16(stakingTime),
		btcutil.Amount(stakingAmout),
		&chaincfg.MainNetParams,
	)
	require.NoError(t, err)
	sti, err := info.TimeLockPathSpendInfo()
	require.NoError(t, err)
	sui, err := info.UnbondingPathSpendInfo()
	require.NoError(t, err)
	ssi, err := info.SlashingPathSpendInfo()
	require.NoError(t, err)

	unbondingInfo, err := btcstaking.BuildUnbondingInfo(
		keysToPubKeys(t, stakerKeys)[0],
		keysToPubKeys(t, finalityKeys),
		keysToPubKeys(t, covenantKeys),
		uint32(covenantQuorum),
		uint16(unbondingTime),
		btcutil.Amount(stakingAmout)-btcutil.Amount(unbondingFee),
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	ubtTx := wire.NewMsgTx(2)
	ubtTx.AddTxIn(wire.NewTxIn(
		wire.NewOutPoint(
			eh,
			0,
		),
		nil,
		nil,
	))
	ubtTx.AddTxOut(unbondingInfo.UnbondingOutput)
	ubtTxHex, err := serializeBTCTxToHex(ubtTx)
	require.NoError(t, err)
	uti, err := unbondingInfo.TimeLockPathSpendInfo()
	require.NoError(t, err)
	usi, err := unbondingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	opReturnOutput := ""
	// if there is more build op_return output
	if len(finalityKeys) == 1 {
		opInfo, err := btcstaking.BuildV0IdentifiableStakingOutputs(
			magicBytes,
			keysToPubKeys(t, stakerKeys)[0],
			keysToPubKeys(t, finalityKeys)[0],
			keysToPubKeys(t, covenantKeys),
			uint32(covenantQuorum),
			uint16(stakingTime),
			btcutil.Amount(stakingAmout),
			&chaincfg.MainNetParams,
		)
		require.NoError(t, err)
		opReturnOutput = hex.EncodeToString(opInfo.OpReturnOutput.PkScript)
	}

	params := Parameters{
		CovenantPublicKeys:         covenantKeys,
		CovenantQuorum:             covenantQuorum,
		FinalityProviderPublicKeys: finalityKeys,
		StakerPublicKey:            stakerKeys[0],
		StakingTime:                stakingTime,
		StakingValue:               stakingAmout,
		StakingTxHash:              eh.String(),
		StakingOutputIndex:         0,
		UnbondingTxVersion:         2,
		UnbondingTime:              unbondingTime,
		UnbondingFee:               unbondingFee,
		MagicBytes:                 hex.EncodeToString(magicBytes),
		Network:                    "mainnet",
	}

	expected := Expected{
		StakingOutputPkScript:              hex.EncodeToString(info.StakingOutput.PkScript),
		StakingOutputValue:                 int(info.StakingOutput.Value),
		StakingTransactionTimeLockScript:   hex.EncodeToString(sti.RevealedLeaf.Script),
		StakingTransactionUnbondingScript:  hex.EncodeToString(sui.RevealedLeaf.Script),
		StakingTransactionSlashingScript:   hex.EncodeToString(ssi.RevealedLeaf.Script),
		UnbondingTransactionHex:            ubtTxHex,
		UnbondingTransactionTimeLockScript: hex.EncodeToString(uti.RevealedLeaf.Script),
		UnbondingTransactionSlashingScript: hex.EncodeToString(usi.RevealedLeaf.Script),
		OpReturnScript:                     opReturnOutput,
	}

	tc := TestCase{
		Description: desc,
		Parameters:  &params,
		Expected:    &expected,
	}

	marshaled, err := json.MarshalIndent(&tc, "", "")
	require.NoError(t, err)
	return string(marshaled)
}
