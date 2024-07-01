package btcstaking

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
)

// buildSlashingTxFromOutpoint builds a valid slashing transaction by creating a new Bitcoin transaction that slashes a portion
// of staked funds and directs them to a specified slashing address. The transaction also includes a change output sent back to
// the specified change address. The slashing rate determines the proportion of staked funds to be slashed.
//
// Parameters:
//   - stakingOutput: The staking output to be spent in the transaction.
//   - stakingAmount: The amount of staked funds in the staking output.
//   - fee: The transaction fee to be paid.
//   - slashingAddress: The Bitcoin address to which the slashed funds will be sent.
//   - changeAddress: The Bitcoin address to receive the change from the transaction.
//   - slashingRate: The rate at which the staked funds will be slashed, expressed as a decimal.
//
// Returns:
//   - *wire.MsgTx: The constructed slashing transaction without a script signature or witness.
//   - error: An error if any validation or construction step fails.
func buildSlashingTxFromOutpoint(
	stakingOutput wire.OutPoint,
	stakingAmount, fee int64,
	slashingAddress, changeAddress btcutil.Address,
	slashingRate sdkmath.LegacyDec,
) (*wire.MsgTx, error) {
	// Validate staking amount
	if stakingAmount <= 0 {
		return nil, fmt.Errorf("staking amount must be larger than 0")
	}

	// Validate slashing rate
	if !IsRateValid(slashingRate) {
		return nil, ErrInvalidSlashingRate
	}

	// Calculate the amount to be slashed
	slashingRateFloat64, err := slashingRate.Float64()
	if err != nil {
		return nil, fmt.Errorf("error converting slashing rate to float64: %w", err)
	}
	slashingAmount := btcutil.Amount(stakingAmount).MulF64(slashingRateFloat64)
	if slashingAmount <= 0 {
		return nil, ErrInsufficientSlashingAmount
	}
	// Generate script for slashing address
	slashingAddrScript, err := txscript.PayToAddrScript(slashingAddress)
	if err != nil {
		return nil, err
	}

	// Calculate the change amount
	changeAmount := btcutil.Amount(stakingAmount) - slashingAmount - btcutil.Amount(fee)
	if changeAmount <= 0 {
		return nil, ErrInsufficientChangeAmount
	}
	// Generate script for change address
	changeAddrScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		return nil, err
	}

	// Create a new btc transaction
	tx := wire.NewMsgTx(wire.TxVersion)
	// TODO: this builds input with sequence number equal to MaxTxInSequenceNum, which
	// means this tx is not replacable.
	input := wire.NewTxIn(&stakingOutput, nil, nil)
	tx.AddTxIn(input)
	tx.AddTxOut(wire.NewTxOut(int64(slashingAmount), slashingAddrScript))
	tx.AddTxOut(wire.NewTxOut(int64(changeAmount), changeAddrScript))

	// Verify that the none of the outputs is a dust output.
	for _, out := range tx.TxOut {
		if mempool.IsDust(out, mempool.DefaultMinRelayTxFee) {
			return nil, ErrDustOutputFound
		}
	}

	return tx, nil
}

func getPossibleStakingOutput(
	stakingTx *wire.MsgTx,
	stakingOutputIdx uint32,
) (*wire.TxOut, error) {
	if stakingTx == nil {
		return nil, fmt.Errorf("provided staking transaction must not be nil")
	}

	if stakingOutputIdx >= uint32(len(stakingTx.TxOut)) {
		return nil, fmt.Errorf("invalid staking output index %d, tx has %d outputs", stakingOutputIdx, len(stakingTx.TxOut))
	}

	stakingOutput := stakingTx.TxOut[stakingOutputIdx]

	if !txscript.IsPayToTaproot(stakingOutput.PkScript) {
		return nil, fmt.Errorf("must be pay to taproot output")
	}

	return stakingOutput, nil
}

// BuildSlashingTxFromStakingTxStrict constructs a valid slashing transaction using information from a staking transaction,
// a specified staking output index, and additional parameters such as slashing and change addresses, transaction fee,
// staking script, script version, and network. This function performs stricter validation compared to BuildSlashingTxFromStakingTx.
//
// Parameters:
//   - stakingTx: The staking transaction from which the staking output is to be used for slashing.
//   - stakingOutputIdx: The index of the staking output in the staking transaction.
//   - slashingAddress: The Bitcoin address to which the slashed funds will be sent.
//   - stakerPk: public key of the staker i.e the btc holder who can spend staking output after lock time
//   - slashingChangeLockTime: lock time which will be used on slashing transaction change output
//   - fee: The transaction fee to be paid.
//   - slashingRate: The rate at which the staked funds will be slashed, expressed as a decimal.
//   - script: The staking script to which the staking output should commit.
//   - net: The network on which transactions should take place (e.g., mainnet, testnet).
//
// Returns:
//   - *wire.MsgTx: The constructed slashing transaction without script signature or witness.
//   - error: An error if any validation or construction step fails.
//
// This function validates the same conditions as BuildSlashingTxFromStakingTx and additionally checks whether the
// staking output at the specified index commits to the provided script and whether the provided script is a valid
// staking script for the given network. If any of these additional validations fail, an error is returned.
func BuildSlashingTxFromStakingTxStrict(
	stakingTx *wire.MsgTx,
	stakingOutputIdx uint32,
	slashingAddress btcutil.Address,
	stakerPk *btcec.PublicKey,
	slashChangeLockTime uint16,
	fee int64,
	slashingRate sdkmath.LegacyDec,
	net *chaincfg.Params,
) (*wire.MsgTx, error) {
	// Get the staking output at the specified index from the staking transaction
	stakingOutput, err := getPossibleStakingOutput(stakingTx, stakingOutputIdx)
	if err != nil {
		return nil, err
	}

	// Create an OutPoint for the staking output
	stakingTxHash := stakingTx.TxHash()
	stakingOutpoint := wire.NewOutPoint(&stakingTxHash, stakingOutputIdx)

	// Create taproot address commiting to timelock script
	si, err := BuildRelativeTimelockTaprootScript(
		stakerPk,
		slashChangeLockTime,
		net,
	)

	if err != nil {
		return nil, err
	}

	// Build slashing tx with the staking output information
	return buildSlashingTxFromOutpoint(
		*stakingOutpoint,
		stakingOutput.Value, fee,
		slashingAddress, si.TapAddress,
		slashingRate)
}

// IsTransferTx Transfer transaction is a transaction which:
// - has exactly one input
// - has exactly one output
func IsTransferTx(tx *wire.MsgTx) error {
	if tx == nil {
		return fmt.Errorf("transfer transaction must have cannot be nil")
	}

	if len(tx.TxIn) != 1 {
		return fmt.Errorf("transfer transaction must have exactly one input")
	}

	if len(tx.TxOut) != 1 {
		return fmt.Errorf("transfer transaction must have exactly one output")
	}

	return nil
}

// IsSimpleTransfer Simple transfer transaction is a transaction which:
// - has exactly one input
// - has exactly one output
// - is not replacable
// - does not have any locktime
func IsSimpleTransfer(tx *wire.MsgTx) error {
	if err := IsTransferTx(tx); err != nil {
		return fmt.Errorf("invalid simple tansfer tx: %w", err)
	}

	if tx.TxIn[0].Sequence != wire.MaxTxInSequenceNum {
		return fmt.Errorf("simple transfer tx must not be replacable")
	}

	if tx.LockTime != 0 {
		return fmt.Errorf("simple transfer tx must not have locktime")
	}
	return nil
}

// validateSlashingTx performs basic checks on a slashing transaction:
// - the slashing transaction is not nil.
// - the slashing transaction has exactly one input.
// - the slashing transaction is non-replaceable.
// - the lock time of the slashing transaction is 0.
// - the slashing transaction has exactly two outputs, and:
//   - the first output must pay to the provided slashing address.
//   - the first output must pay at least (staking output value * slashing rate) to the slashing address.
//   - neither of the outputs are considered dust.
//
// - the min fee for slashing tx is preserved
func validateSlashingTx(
	slashingTx *wire.MsgTx,
	slashingAddress btcutil.Address,
	slashingRate sdkmath.LegacyDec,
	slashingTxMinFee, stakingOutputValue int64,
	stakerPk *btcec.PublicKey,
	slashingChangeLockTime uint16,
	net *chaincfg.Params,
) error {
	// Verify that the slashing transaction is not nil.
	if slashingTx == nil {
		return fmt.Errorf("slashing transaction must not be nil")
	}

	// Verify that the slashing transaction has exactly one input.
	if len(slashingTx.TxIn) != 1 {
		return fmt.Errorf("slashing transaction must have exactly one input")
	}

	// Verify that the slashing transaction is non-replaceable.
	if slashingTx.TxIn[0].Sequence != wire.MaxTxInSequenceNum {
		return fmt.Errorf("slashing transaction must not be replaceable")
	}

	// Verify that lock time of the slashing transaction is 0.
	if slashingTx.LockTime != 0 {
		return fmt.Errorf("slashing tx must not have locktime")
	}

	// Verify that the slashing transaction has exactly two outputs.
	if len(slashingTx.TxOut) != 2 {
		return fmt.Errorf("slashing transaction must have exactly 2 outputs")
	}

	// Verify that at least staking output value * slashing rate is slashed.
	slashingRateFloat64, err := slashingRate.Float64()
	if err != nil {
		return fmt.Errorf("error converting slashing rate to float64: %w", err)
	}
	minSlashingAmount := btcutil.Amount(stakingOutputValue).MulF64(slashingRateFloat64)
	if btcutil.Amount(slashingTx.TxOut[0].Value) < minSlashingAmount {
		return fmt.Errorf("slashing transaction must slash at least staking output value * slashing rate")
	}

	// Verify that the first output pays to the provided slashing address.
	slashingPkScript, err := txscript.PayToAddrScript(slashingAddress)
	if err != nil {
		return fmt.Errorf("error creating slashing pk script: %w", err)
	}
	if !bytes.Equal(slashingTx.TxOut[0].PkScript, slashingPkScript) {
		return fmt.Errorf("slashing transaction must pay to the provided slashing address")
	}

	// Verify that the second output pays to the taproot address which locks funds for
	// slashingChangeLockTime
	si, err := BuildRelativeTimelockTaprootScript(
		stakerPk,
		slashingChangeLockTime,
		net,
	)

	if err != nil {
		return fmt.Errorf("error creating change timelock script: %w", err)
	}

	if !bytes.Equal(slashingTx.TxOut[1].PkScript, si.PkScript) {
		return fmt.Errorf("invalid slashing tx change output pkscript, expected: %s, got: %s", hex.EncodeToString(si.PkScript), hex.EncodeToString(slashingTx.TxOut[1].PkScript))
	}

	// Verify that the none of the outputs is a dust output.
	for _, out := range slashingTx.TxOut {
		if mempool.IsDust(out, mempool.DefaultMinRelayTxFee) {
			return ErrDustOutputFound
		}
	}

	/*
		Check Fees
	*/
	// Check that values of slashing and staking transaction are larger than 0
	if slashingTx.TxOut[0].Value <= 0 || stakingOutputValue <= 0 {
		return fmt.Errorf("values of slashing and staking transaction must be larger than 0")
	}

	// Calculate the sum of output values in the slashing transaction.
	slashingTxOutSum := int64(0)
	for _, out := range slashingTx.TxOut {
		slashingTxOutSum += out.Value
	}

	// Ensure that the staking transaction value is larger than the sum of slashing transaction output values.
	if stakingOutputValue <= slashingTxOutSum {
		return fmt.Errorf("slashing transaction must not spend more than staking transaction")
	}

	// Ensure that the slashing transaction fee is larger than the specified minimum fee.
	if stakingOutputValue-slashingTxOutSum < slashingTxMinFee {
		return fmt.Errorf("slashing transaction fee must be larger than %d", slashingTxMinFee)
	}

	return nil
}

// CheckTransactions validates all relevant data of slashing and funding transaction.
// - both transactions are valid from pov of BTC rules
// - funding transaction has output committing to the provided script
// - slashing transaction is valid
// - slashing transaction input hash is pointing to funding transaction hash
// - slashing transaction input index is pointing to funding transaction output commiting to the script
func CheckTransactions(
	slashingTx *wire.MsgTx,
	fundingTransaction *wire.MsgTx,
	fundingOutputIdx uint32,
	slashingTxMinFee int64,
	slashingRate sdkmath.LegacyDec,
	slashingAddress btcutil.Address,
	stakerPk *btcec.PublicKey,
	slashingChangeLockTime uint16,
	net *chaincfg.Params,
) error {
	if slashingTx == nil || fundingTransaction == nil {
		return fmt.Errorf("slashing and funding transactions must not be nil")
	}

	if err := blockchain.CheckTransactionSanity(btcutil.NewTx(slashingTx)); err != nil {
		return fmt.Errorf("slashing transaction does not obey BTC rules: %w", err)
	}

	if err := blockchain.CheckTransactionSanity(btcutil.NewTx(fundingTransaction)); err != nil {
		return fmt.Errorf("funding transaction does not obey BTC rules: %w", err)
	}

	// Check if slashing tx min fee is valid
	if slashingTxMinFee <= 0 {
		return fmt.Errorf("slashing transaction min fee must be larger than 0")
	}

	// Check if slashing rate is in the valid range (0,1)
	if !IsRateValid(slashingRate) {
		return ErrInvalidSlashingRate
	}

	if fundingOutputIdx >= uint32(len(fundingTransaction.TxOut)) {
		return fmt.Errorf("invalid funding output index %d, tx has %d outputs", fundingOutputIdx, len(fundingTransaction.TxOut))
	}

	stakingOutput := fundingTransaction.TxOut[fundingOutputIdx]
	// 3. Check if slashing transaction is valid
	if err := validateSlashingTx(
		slashingTx,
		slashingAddress,
		slashingRate,
		slashingTxMinFee,
		stakingOutput.Value,
		stakerPk,
		slashingChangeLockTime,
		net); err != nil {
		return err
	}

	// 4. Check that slashing transaction input is pointing to staking transaction
	stakingTxHash := fundingTransaction.TxHash()
	if !slashingTx.TxIn[0].PreviousOutPoint.Hash.IsEqual(&stakingTxHash) {
		return fmt.Errorf("slashing transaction must spend staking output")
	}

	// 5. Check that index of the fund output matches index of the input in slashing transaction
	if slashingTx.TxIn[0].PreviousOutPoint.Index != fundingOutputIdx {
		return fmt.Errorf("slashing transaction input must spend staking output")
	}
	return nil
}

func signTxWithOneScriptSpendInputFromTapLeafInternal(
	txToSign *wire.MsgTx,
	fundingOutput *wire.TxOut,
	privKey *btcec.PrivateKey,
	tapLeaf txscript.TapLeaf) (*schnorr.Signature, error) {

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		fundingOutput.PkScript,
		fundingOutput.Value,
	)

	sigHashes := txscript.NewTxSigHashes(txToSign, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		txToSign, sigHashes, 0, fundingOutput.Value,
		fundingOutput.PkScript, tapLeaf, txscript.SigHashDefault,
		privKey,
	)

	if err != nil {
		return nil, err
	}

	parsedSig, err := schnorr.ParseSignature(sig)

	if err != nil {
		return nil, err
	}

	return parsedSig, nil
}

// SignTxWithOneScriptSpendInputFromTapLeaf signs transaction with one input coming
// from script spend output.
// It does not do any validations, expect that txToSign has exactly one input.
func SignTxWithOneScriptSpendInputFromTapLeaf(
	txToSign *wire.MsgTx,
	fundingOutput *wire.TxOut,
	privKey *btcec.PrivateKey,
	tapLeaf txscript.TapLeaf,
) (*schnorr.Signature, error) {
	if txToSign == nil {
		return nil, fmt.Errorf("tx to sign must not be nil")
	}

	if fundingOutput == nil {
		return nil, fmt.Errorf("funding output must not be nil")
	}

	if privKey == nil {
		return nil, fmt.Errorf("private key must not be nil")
	}

	if len(txToSign.TxIn) != 1 {
		return nil, fmt.Errorf("tx to sign must have exactly one input")
	}

	return signTxWithOneScriptSpendInputFromTapLeafInternal(
		txToSign,
		fundingOutput,
		privKey,
		tapLeaf,
	)
}

// SignTxWithOneScriptSpendInputFromScript signs transaction with one input coming
// from script spend output with provided script.
// It does not do any validations, expect that txToSign has exactly one input.
func SignTxWithOneScriptSpendInputFromScript(
	txToSign *wire.MsgTx,
	fundingOutput *wire.TxOut,
	privKey *btcec.PrivateKey,
	script []byte,
) (*schnorr.Signature, error) {
	tapLeaf := txscript.NewBaseTapLeaf(script)
	return SignTxWithOneScriptSpendInputFromTapLeaf(txToSign, fundingOutput, privKey, tapLeaf)
}

// SignTxWithOneScriptSpendInputStrict signs transaction with one input coming
// from script spend output with provided script.
// It checks:
// - txToSign is not nil
// - txToSign has exactly one input
// - fundingTx is not nil
// - fundingTx has one output committing to the provided script
// - txToSign input is pointing to the correct output in fundingTx
func SignTxWithOneScriptSpendInputStrict(
	txToSign *wire.MsgTx,
	fundingTx *wire.MsgTx,
	fundingOutputIdx uint32,
	signedScriptPath []byte,
	privKey *btcec.PrivateKey,
) (*schnorr.Signature, error) {

	if err := checkTxBeforeSigning(txToSign, fundingTx, fundingOutputIdx); err != nil {
		return nil, fmt.Errorf("invalid tx: %w", err)
	}

	fundingOutput := fundingTx.TxOut[fundingOutputIdx]

	return SignTxWithOneScriptSpendInputFromScript(txToSign, fundingOutput, privKey, signedScriptPath)
}

// EncSignTxWithOneScriptSpendInputStrict is encrypted version of
// SignTxWithOneScriptSpendInputStrict with the output to be encrypted
// by an encryption key (adaptor signature)
func EncSignTxWithOneScriptSpendInputStrict(
	txToSign *wire.MsgTx,
	fundingTx *wire.MsgTx,
	fundingOutputIdx uint32,
	signedScriptPath []byte,
	privKey *btcec.PrivateKey,
	encKey *asig.EncryptionKey,
) (*asig.AdaptorSignature, error) {

	if err := checkTxBeforeSigning(txToSign, fundingTx, fundingOutputIdx); err != nil {
		return nil, fmt.Errorf("invalid tx: %w", err)
	}

	fundingOutput := fundingTx.TxOut[fundingOutputIdx]

	sigHash, err := getSigHash(txToSign, fundingOutput, signedScriptPath)
	if err != nil {
		return nil, err
	}

	adaptorSig, err := asig.EncSign(privKey, encKey, sigHash)
	if err != nil {
		return nil, err
	}

	return adaptorSig, nil
}

func checkTxBeforeSigning(txToSign *wire.MsgTx, fundingTx *wire.MsgTx, fundingOutputIdx uint32) error {
	if txToSign == nil {
		return fmt.Errorf("tx to sign must not be nil")
	}

	if len(txToSign.TxIn) != 1 {
		return fmt.Errorf("tx to sign must have exactly one input")
	}

	if fundingOutputIdx >= uint32(len(fundingTx.TxOut)) {
		return fmt.Errorf("invalid funding output index %d, tx has %d outputs", fundingOutputIdx, len(fundingTx.TxOut))
	}

	fundingTxHash := fundingTx.TxHash()

	if !txToSign.TxIn[0].PreviousOutPoint.Hash.IsEqual(&fundingTxHash) {
		return fmt.Errorf("txToSign must input point to fundingTx")
	}

	if txToSign.TxIn[0].PreviousOutPoint.Index != fundingOutputIdx {
		return fmt.Errorf("txToSign inpunt index must point to output with provided script")
	}

	return nil
}

// getSigHash returns the sig hash of the given tx spending the given tx output
// via the given script path
// signatures over this tx have to be signed over the message being the sig hash
func getSigHash(transaction *wire.MsgTx, fundingOutput *wire.TxOut, script []byte) ([]byte, error) {
	if fundingOutput == nil {
		return nil, fmt.Errorf("funding output must not be nil")
	}

	if transaction == nil {
		return nil, fmt.Errorf("tx to verify not be nil")
	}

	if len(transaction.TxIn) != 1 {
		return nil, fmt.Errorf("tx to sign must have exactly one input")
	}

	tapLeaf := txscript.NewBaseTapLeaf(script)

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		fundingOutput.PkScript,
		fundingOutput.Value,
	)

	sigHashes := txscript.NewTxSigHashes(transaction, inputFetcher)

	return txscript.CalcTapscriptSignaturehash(
		sigHashes, txscript.SigHashDefault, transaction, 0, inputFetcher, tapLeaf,
	)
}

// VerifyTransactionSigWithOutput verifies that:
// - provided transaction has exactly one input
// - provided signature is valid schnorr BIP340 signature
// - provided signature is signing whole provided transaction	(SigHashDefault)
func VerifyTransactionSigWithOutput(
	transaction *wire.MsgTx,
	fundingOutput *wire.TxOut,
	script []byte,
	pubKey *btcec.PublicKey,
	signature []byte,
) error {
	if pubKey == nil {
		return fmt.Errorf("public key must not be nil")
	}

	sigHash, err := getSigHash(transaction, fundingOutput, script)
	if err != nil {
		return err
	}

	parsedSig, err := schnorr.ParseSignature(signature)
	if err != nil {
		return err
	}

	if !parsedSig.Verify(sigHash, pubKey) {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

// EncVerifyTransactionSigWithOutput verifies that:
// - provided transaction has exactly one input
// - provided signature is valid adaptor signature
// - provided signature is signing whole provided transaction (SigHashDefault)
func EncVerifyTransactionSigWithOutput(
	transaction *wire.MsgTx,
	fundingOut *wire.TxOut,
	script []byte,
	pubKey *btcec.PublicKey,
	encKey *asig.EncryptionKey,
	signature *asig.AdaptorSignature,
) error {
	if pubKey == nil {
		return fmt.Errorf("public key must not be nil")
	}
	if encKey == nil {
		return fmt.Errorf("encryption key must not be nil")
	}

	sigHash, err := getSigHash(transaction, fundingOut, script)
	if err != nil {
		return err
	}

	return signature.EncVerify(pubKey, encKey, sigHash)
}
