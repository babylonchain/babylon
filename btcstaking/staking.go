package btcstaking

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// we expect signatures from 3 signers
	expectedMultiSigSigners = 3

	// Point with unknown discrete logarithm defined in: https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki#constructing-and-spending-taproot-outputs
	// using it as internal public key efectively disables taproot key spends
	unspendableKeyPath = "0250929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0"
)

var (
	unspendableKeyPathKey = unspendableKeyPathInternalPubKeyInternal(unspendableKeyPath)
)

func unspendableKeyPathInternalPubKeyInternal(keyHex string) btcec.PublicKey {
	keyBytes, err := hex.DecodeString(keyHex)

	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}

	// We are using btcec here, as key is 33 byte compressed format.
	pubKey, err := btcec.ParsePubKey(keyBytes)

	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}
	return *pubKey
}

// Following methods are copied from btcd. In most recent they are not exported.
// TODO: on btcd master those are already exported. Remove this copies
// when this will be released.
func isSmallInt(op byte) bool {
	return op == txscript.OP_0 || (op >= txscript.OP_1 && op <= txscript.OP_16)
}

func asSmallInt(op byte) int {
	if op == txscript.OP_0 {
		return 0
	}

	return int(op - (txscript.OP_1 - 1))
}

func checkMinimalDataEncoding(v []byte) error {
	if len(v) == 0 {
		return nil
	}

	// Check that the number is encoded with the minimum possible
	// number of bytes.
	//
	// If the most-significant-byte - excluding the sign bit - is zero
	// then we're not minimal.  Note how this test also rejects the
	// negative-zero encoding, [0x80].
	if v[len(v)-1]&0x7f == 0 {
		// One exception: if there's more than one byte and the most
		// significant bit of the second-most-significant-byte is set
		// it would conflict with the sign bit.  An example of this case
		// is +-255, which encode to 0xff00 and 0xff80 respectively.
		// (big-endian).
		if len(v) == 1 || v[len(v)-2]&0x80 == 0 {
			return fmt.Errorf("numeric value encoded as %x is "+
				"not minimally encoded", v)
		}
	}

	return nil
}

func makeScriptNum(v []byte, requireMinimal bool, scriptNumLen int) (int64, error) {
	// Interpreting data requires that it is not larger than
	// the the passed scriptNumLen value.
	if len(v) > scriptNumLen {
		return 0, fmt.Errorf("numeric value encoded as %x is %d bytes "+
			"which exceeds the max allowed of %d", v, len(v),
			scriptNumLen)
	}

	// Enforce minimal encoded if requested.
	if requireMinimal {
		if err := checkMinimalDataEncoding(v); err != nil {
			return 0, err
		}
	}

	// Zero is encoded as an empty byte slice.
	if len(v) == 0 {
		return 0, nil
	}

	// Decode from little endian.
	var result int64
	for i, val := range v {
		result |= int64(val) << uint8(8*i)
	}

	// When the most significant byte of the input bytes has the sign bit
	// set, the result is negative.  So, remove the sign bit from the result
	// and make it negative.
	if v[len(v)-1]&0x80 != 0 {
		// The maximum length of v has already been determined to be 4
		// above, so uint8 is enough to cover the max possible shift
		// value of 24.
		result &= ^(int64(0x80) << uint8(8*(len(v)-1)))
		return -result, nil
	}

	return result, nil
}

// StakingScriptData is a struct that holds data parsed from staking script
type StakingScriptData struct {
	StakerKey    *btcec.PublicKey
	ValidatorKey *btcec.PublicKey
	CovenantKey  *btcec.PublicKey
	StakingTime  uint16
}

// StakingOutputInfo holds info about whole staking output:
// - data derived from the script
// - staking amount in staking output
// - staking pk script
type StakingOutputInfo struct {
	StakingScriptData *StakingScriptData
	StakingAmount     btcutil.Amount
	StakingPkScript   []byte
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

// BuildStakingScript builds a staking script in the following format:
// <StakerKey> OP_CHECKSIG
// OP_NOTIF
//
//	<StakerKey> OP_CHECKSIG <ValidatorKey> OP_CHECKSIGADD <CovenantKey> OP_CHECKSIGADD 3 OP_NUMEQUAL
//
// OP_ELSE
//
//	<stTime> OP_CHECKSEQUENCEVERIFY
//
// OP_ENDIF
func (sd *StakingScriptData) BuildStakingScript() ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(sd.StakerKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_NOTIF)
	builder.AddData(schnorr.SerializePubKey(sd.StakerKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddData(schnorr.SerializePubKey(sd.ValidatorKey))
	builder.AddOp(txscript.OP_CHECKSIGADD)
	builder.AddData(schnorr.SerializePubKey(sd.CovenantKey))
	builder.AddOp(txscript.OP_CHECKSIGADD)
	builder.AddInt64(expectedMultiSigSigners)
	builder.AddOp(txscript.OP_NUMEQUAL)
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(int64(sd.StakingTime))
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_ENDIF)
	return builder.Script()

}

// ParseStakingTransactionScript parses provided script. If script is not a valid staking script
// error is returned. If script is valid, StakingScriptData is returned, which contains all
// relevant data parsed from the script.
// Only stateless checks are performed.
func ParseStakingTransactionScript(script []byte) (*StakingScriptData, error) {
	// <StakerKey> OP_CHECKSIG
	// OP_NOTIF
	//
	//	<StakerKey> OP_CHECKSIG <ValidatorKey> OP_CHECKSIGADD <CovenantKey> OP_CHECKSIGADD 3 OP_NUMEQUAL
	//
	// OP_ELSE
	//
	//	<stTime> OP_CHECKSEQUENCEVERIFY
	//
	// OP_ENDIF
	type templateMatch struct {
		expectCanonicalInt bool
		maxIntBytes        int
		opcode             byte
		extractedInt       int64
		extractedData      []byte
	}
	var template = [15]templateMatch{
		{opcode: txscript.OP_DATA_32},
		{opcode: txscript.OP_CHECKSIG},
		{opcode: txscript.OP_NOTIF},
		{opcode: txscript.OP_DATA_32},
		{opcode: txscript.OP_CHECKSIG},
		{opcode: txscript.OP_DATA_32},
		{opcode: txscript.OP_CHECKSIGADD},
		{opcode: txscript.OP_DATA_32},
		{opcode: txscript.OP_CHECKSIGADD},
		{expectCanonicalInt: true, maxIntBytes: 4},
		{opcode: txscript.OP_NUMEQUAL},
		{opcode: txscript.OP_ELSE},
		{expectCanonicalInt: true, maxIntBytes: 4},
		{opcode: txscript.OP_CHECKSEQUENCEVERIFY},
		{opcode: txscript.OP_ENDIF},
	}

	var templateOffset int
	tokenizer := txscript.MakeScriptTokenizer(0, script)

	for tokenizer.Next() {
		// Not an staking script if it has more opcodes than expected in the
		// template.
		if templateOffset >= len(template) {
			return nil, nil
		}

		op := tokenizer.Opcode()
		data := tokenizer.Data()
		tplEntry := &template[templateOffset]
		if tplEntry.expectCanonicalInt {
			switch {
			case data != nil:
				val, err := makeScriptNum(data, true, tplEntry.maxIntBytes)
				if err != nil {
					return nil, err
				}
				tplEntry.extractedInt = int64(val)

			case isSmallInt(op):
				tplEntry.extractedInt = int64(asSmallInt(op))

			// Not an staking script if this is not int
			default:
				return nil, nil
			}
		} else {
			if op != tplEntry.opcode {
				return nil, nil
			}

			tplEntry.extractedData = data
		}

		templateOffset++
	}
	if err := tokenizer.Err(); err != nil {
		return nil, err
	}
	if !tokenizer.Done() || templateOffset != len(template) {
		return nil, nil
	}

	// At this point, the script appears to be an valid staking script. Extract relevant data and perform
	// some initial validations.

	// Staker public key from the path without multisig i.e path where sats are locked
	// for staking duration
	stakerPk1, err := schnorr.ParsePubKey(template[0].extractedData)
	if err != nil {
		return nil, err
	}

	// Staker public key from the path with multisig
	if _, err := schnorr.ParsePubKey(template[3].extractedData); err != nil {
		return nil, err
	}

	if !bytes.Equal(template[0].extractedData, template[3].extractedData) {
		return nil, fmt.Errorf("staker public key on lock path and multisig path are different")
	}

	// Delegator public key
	validatorPk, err := schnorr.ParsePubKey(template[5].extractedData)

	if err != nil {
		return nil, err
	}

	// Covenant public key
	covenantPk, err := schnorr.ParsePubKey(template[7].extractedData)

	if err != nil {
		return nil, err
	}

	// validate number of mulitsig signers
	if template[9].extractedInt != expectedMultiSigSigners {
		return nil, fmt.Errorf("expected %d multisig signers, got %d", expectedMultiSigSigners, template[9].extractedInt)
	}

	// validate staking time
	if template[12].extractedInt < 0 || template[12].extractedInt > math.MaxUint16 {
		return nil, fmt.Errorf("invalid staking time %d", template[12].extractedInt)
	}

	// we do not need to check error here, as we already validated that all public keys are not nil
	scriptData, err := NewStakingScriptData(
		stakerPk1,
		validatorPk,
		covenantPk,
		uint16(template[12].extractedInt),
	)

	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}

	return scriptData, nil
}

func UnspendableKeyPathInternalPubKey() btcec.PublicKey {
	return unspendableKeyPathKey
}

// TaprootAddressForScript returns a Taproot address commiting to the given script, built taproot tree
// has only one leaf node.
func TaprootAddressForScript(
	script []byte,
	internalPubKey *btcec.PublicKey,
	net *chaincfg.Params) (*btcutil.AddressTaproot, error) {

	tapLeaf := txscript.NewBaseTapLeaf(script)

	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	tapScriptRootHash := tapScriptTree.RootNode.TapHash()

	outputKey := txscript.ComputeTaprootOutputKey(
		internalPubKey, tapScriptRootHash[:],
	)

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(outputKey), net)

	if err != nil {
		return nil, fmt.Errorf("error encoding Taproot address: %v", err)
	}

	return address, nil
}

// BuildUnspendableTaprootPkScript builds taproot pkScript which commits to the provided script with
// unspendable spending key path.
func BuildUnspendableTaprootPkScript(rawScript []byte, net *chaincfg.Params) ([]byte, error) {
	internalPubKey := UnspendableKeyPathInternalPubKey()

	address, err := TaprootAddressForScript(rawScript, &internalPubKey, net)

	if err != nil {
		return nil, err
	}

	pkScript, err := txscript.PayToAddrScript(address)

	if err != nil {
		return nil, err
	}

	return pkScript, nil
}

// BuildStakingOutput builds out which is necessary for staking transaction to stake funds.
func BuildStakingOutput(
	stakerKey,
	validatorKey,
	covenantKey *btcec.PublicKey,
	stTime uint16,
	stAmount btcutil.Amount,
	net *chaincfg.Params) (*wire.TxOut, []byte, error) {

	sd, err := NewStakingScriptData(stakerKey, validatorKey, covenantKey, stTime)

	if err != nil {
		return nil, nil, err
	}

	script, err := sd.BuildStakingScript()

	if err != nil {
		return nil, nil, err
	}

	pkScript, err := BuildUnspendableTaprootPkScript(script, net)

	if err != nil {
		return nil, nil, err
	}

	return wire.NewTxOut(int64(stAmount), pkScript), script, nil
}

// NewWitnessFromStakingScriptAndSignature creates witness for spending
// staking from the given staking script and the given signature of
// a single party in the multisig
func NewWitnessFromStakingScriptAndSignature(
	stakingScript []byte,
	sig *schnorr.Signature,
) (wire.TxWitness, error) {
	// get ctrlBlockBytes
	internalPubKey := UnspendableKeyPathInternalPubKey()
	tapLeaf := txscript.NewBaseTapLeaf(stakingScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(&internalPubKey)
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	// compose witness stack
	witnessStack := wire.TxWitness(make([][]byte, 3))
	witnessStack[0] = sig.Serialize()
	witnessStack[1] = stakingScript
	witnessStack[2] = ctrlBlockBytes
	return witnessStack, nil
}

// BuildWitnessToSpendStakingOutput builds witness for spending staking as single staker
// Current assumptions:
// - staking output is the only input to the transaction
// - staking output is valid staking output
func BuildWitnessToSpendStakingOutput(
	slashingMsgTx *wire.MsgTx, // slashing tx
	stakingOutput *wire.TxOut,
	stakingScript []byte,
	privKey *btcec.PrivateKey,
) (wire.TxWitness, error) {
	tapLeaf := txscript.NewBaseTapLeaf(stakingScript)
	sig, err := SignTxWithOneScriptSpendInputFromTapLeaf(slashingMsgTx, stakingOutput, privKey, tapLeaf)
	if err != nil {
		return nil, err
	}

	return NewWitnessFromStakingScriptAndSignature(stakingScript, sig)
}

// ValidateStakingOutputPkScript validates that:
// - provided output commits to the given script with unspendable spending key path
// - provided script is valid staking script
func ValidateStakingOutputPkScript(
	output *wire.TxOut,
	script []byte,
	net *chaincfg.Params) (*StakingScriptData, error) {
	if output == nil {
		return nil, fmt.Errorf("provided output cannot be nil")
	}

	pkScript, err := BuildUnspendableTaprootPkScript(script, net)

	if err != nil {
		return nil, err
	}

	if !bytes.Equal(output.PkScript, pkScript) {
		return nil, fmt.Errorf("output does not commit to the given script")
	}

	return ParseStakingTransactionScript(script)
}

// BuildSlashingTxFromOutpoint builds valid slashing transaction, using provided:
// - stakingOutput - staking output
// - slashingAddress - address to which slashed funds will go
// - fee - fee for the transaction
// It does not attach script sig to the transaction nor the witness.
// It only validates that provided address is standard btc address and slashing value is larger than 0
func BuildSlashingTxFromOutpoint(
	stakingOutput wire.OutPoint,
	slashingAddress btcutil.Address,
	slashingValue int64) (*wire.MsgTx, error) {

	addrScript, err := txscript.PayToAddrScript(slashingAddress)

	if err != nil {
		return nil, err
	}

	if slashingValue <= 0 {
		return nil, fmt.Errorf("slashing value cannot be smaller or equal 0")
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	// TODO: this builds input with sequence number equal to MaxTxInSequenceNum, which
	// means this tx is not replacable.
	input := wire.NewTxIn(&stakingOutput, nil, nil)
	tx.AddTxIn(input)
	tx.AddTxOut(wire.NewTxOut(slashingValue, addrScript))
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

// BuildSlashingTxFromOutpoint builds valid slashing transaction, using provided:
// - stakingTx - staking trasaction
// - stakingOutputIdx - index of the output committing to staking script
// - slashingAddress - address to which slashed funds will go
// - fee - fee for the transaction
// It does not attach script sig to the transaction nor the witness.
// It validates:
// - stakingTx is not nil
// - staking tx has output at stakingOutputIdx
// - staking output at stakingOutputIdx is valid staking output i.e p2tr output
func BuildSlashingTxFromStakingTx(
	stakingTx *wire.MsgTx,
	stakingOutputIdx uint32,
	slashingAddress btcutil.Address,
	fee int64,
) (*wire.MsgTx, error) {
	stakingOutput, err := getPossibleStakingOutput(stakingTx, stakingOutputIdx)

	if err != nil {
		return nil, err
	}

	stakingTxHash := stakingTx.TxHash()

	stakingOutpoint := wire.NewOutPoint(&stakingTxHash, stakingOutputIdx)

	return BuildSlashingTxFromOutpoint(*stakingOutpoint, slashingAddress, stakingOutput.Value-fee)
}

// BuildSlashingTxFromStakingTxStrict builds valid slashing transaction, using provided:
// - stakingTx - staking trasaction
// - stakingOutputIdx - index of the output committing to staking script
// - slashingAddress - address to which slashed funds will go
// - fee - fee for the transaction
// - script - staking script to which staking output should commit
// - scriptVersion - version of the script
// - net - network on wchich transactions should take place
// It validates:
// - the same stuff as BuildSlashingTxFromStakingTx
// - wheter staking output commits to the provided script
// - whether provided script is valid staking script
func BuildSlashingTxFromStakingTxStrict(
	stakingTx *wire.MsgTx,
	stakingOutputIdx uint32,
	slashingAddress btcutil.Address,
	fee int64,
	script []byte,
	net *chaincfg.Params,
) (*wire.MsgTx, error) {
	stakingOutput, err := getPossibleStakingOutput(stakingTx, stakingOutputIdx)

	if err != nil {
		return nil, err
	}

	if _, err := ValidateStakingOutputPkScript(stakingOutput, script, net); err != nil {
		return nil, err
	}

	stakingTxHash := stakingTx.TxHash()

	stakingOutpoint := wire.NewOutPoint(&stakingTxHash, stakingOutputIdx)

	return BuildSlashingTxFromOutpoint(*stakingOutpoint, slashingAddress, stakingOutput.Value-fee)
}

// Transfer transaction is a transaction which:
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

// Simple transfer transaction is a transaction which:
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

// IsSlashingTx perform basic checks on slashing transaction:
// - slashing transaction is not nil
// - slashing transaction has exactly one input
// - slashing transaction is not replacable
// - slashing transaction has exactly one output
// - slashing transaction locktime is 0
// - slashing transaction output is simple pay to address script paying to provided slashing address
func IsSlashingTx(slashingTx *wire.MsgTx, slashingAddress btcutil.Address) error {
	if err := IsSimpleTransfer(slashingTx); err != nil {
		return fmt.Errorf("invalid slashing tx: %w", err)
	}

	pkScript, err := txscript.PayToAddrScript(slashingAddress)

	if err != nil {
		return err
	}

	if !bytes.Equal(slashingTx.TxOut[0].PkScript, pkScript) {
		return fmt.Errorf("slashing transaction must pay to provided slashing address")
	}

	return nil
}

// GetIdxOutputCommitingToScript retrieves index of the output committing to the provided script.
// It returns error if:
// - tx is nil
// - tx does not have output committing to the provided script
// - tx has more than one output committing to the provided script
func GetIdxOutputCommitingToScript(
	tx *wire.MsgTx,
	script []byte,
	net *chaincfg.Params) (int, error) {

	if tx == nil {
		return -1, fmt.Errorf("provided staking transaction must not be nil")
	}

	script, err := BuildUnspendableTaprootPkScript(script, net)

	if err != nil {
		return -1, err
	}

	var comittingOutputIdx int = -1
	for i, out := range tx.TxOut {
		if bytes.Equal(out.PkScript, script) && comittingOutputIdx < 0 {
			comittingOutputIdx = i
		} else if bytes.Equal(out.PkScript, script) && comittingOutputIdx >= 0 {
			return -1, fmt.Errorf("transaction has more than one output committing to the provided script")
		}
	}

	if comittingOutputIdx < 0 {
		return -1, fmt.Errorf("transaction does not have output committing to the provided script")
	}
	return comittingOutputIdx, nil
}

// CheckTransactions validates all relevant data of slashing and funding transaction.
// - slashing transaction is valid
// - funding transaction script is valid
// - funding transaction has output committing to the provided script
// - slashing transaction input is pointing to funding transaction output commiting to the script
// - that min fee for slashing tx is preserved
// In case of success, it returns data extracted from valid staking script and staking amount.
func CheckTransactions(
	slashingTx *wire.MsgTx,
	fundingTransaction *wire.MsgTx,
	slashingTxMinFee int64,
	slashingAddress btcutil.Address,
	script []byte,
	net *chaincfg.Params,
) (*StakingOutputInfo, error) {
	if slashingTxMinFee <= 0 {
		return nil, fmt.Errorf("slashing transaction min fee must be larger than 0")
	}

	// 1. Check slashing tx
	if err := IsSlashingTx(slashingTx, slashingAddress); err != nil {
		return nil, err
	}

	// 2. Check staking script.
	scriptData, err := ParseStakingTransactionScript(script)

	if err != nil {
		return nil, err
	}

	// 3. Check that staking transaction has output committing to the provided script
	stakingOutputIdx, err := GetIdxOutputCommitingToScript(fundingTransaction, script, net)

	if err != nil {
		return nil, err
	}

	// 4. Check that slashing transaction input is pointing to staking transaction
	stakingTxHash := fundingTransaction.TxHash()
	if !slashingTx.TxIn[0].PreviousOutPoint.Hash.IsEqual(&stakingTxHash) {
		return nil, fmt.Errorf("slashing transaction must spend staking output")
	}

	// 5. Check that index of the fund output matches index of the input in slashing transaction
	if slashingTx.TxIn[0].PreviousOutPoint.Index != uint32(stakingOutputIdx) {
		return nil, fmt.Errorf("slashing transaction input must spend staking output")
	}

	stakingOutput := fundingTransaction.TxOut[stakingOutputIdx]

	// 6. Check fees
	if slashingTx.TxOut[0].Value <= 0 || stakingOutput.Value <= 0 {
		return nil, fmt.Errorf("values of slashing and staking transaction must be larger than 0")
	}

	if stakingOutput.Value <= slashingTx.TxOut[0].Value {
		return nil, fmt.Errorf("slashing transaction must not spend more than staking transaction")
	}

	if stakingOutput.Value-slashingTx.TxOut[0].Value < slashingTxMinFee {
		return nil, fmt.Errorf("slashing transaction fee must be larger than %d", slashingTxMinFee)
	}

	return &StakingOutputInfo{
		StakingScriptData: scriptData,
		StakingAmount:     btcutil.Amount(stakingOutput.Value),
		StakingPkScript:   stakingOutput.PkScript,
	}, nil
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
	privKey *btcec.PrivateKey,
	script []byte,
	net *chaincfg.Params,
) (*schnorr.Signature, error) {

	if txToSign == nil {
		return nil, fmt.Errorf("tx to sign must not be nil")
	}

	if len(txToSign.TxIn) != 1 {
		return nil, fmt.Errorf("tx to sign must have exactly one input")
	}

	scriptIdx, err := GetIdxOutputCommitingToScript(fundingTx, script, net)

	if err != nil {
		return nil, err
	}

	fundingTxHash := fundingTx.TxHash()

	if !txToSign.TxIn[0].PreviousOutPoint.Hash.IsEqual(&fundingTxHash) {
		return nil, fmt.Errorf("txToSign must input point to fundingTx")
	}

	if txToSign.TxIn[0].PreviousOutPoint.Index != uint32(scriptIdx) {
		return nil, fmt.Errorf("txToSign inpunt index must point to output with provided script")
	}

	fundingOutput := fundingTx.TxOut[scriptIdx]

	return SignTxWithOneScriptSpendInputFromScript(txToSign, fundingOutput, privKey, script)
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
	signature []byte) error {

	if fundingOutput == nil {
		return fmt.Errorf("funding output must not be nil")
	}

	return VerifyTransactionSigWithOutputData(
		transaction,
		fundingOutput.PkScript,
		fundingOutput.Value,
		script,
		pubKey,
		signature,
	)
}

// VerifyTransactionSigWithOutputData verifies that:
// - provided transaction has exactly one input
// - provided signature is valid schnorr BIP340 signature
// - provided signature is signing whole provided transaction	(SigHashDefault)
func VerifyTransactionSigWithOutputData(
	transaction *wire.MsgTx,
	fundingOutputPkScript []byte,
	fundingOutputValue int64,
	script []byte,
	pubKey *btcec.PublicKey,
	signature []byte) error {

	if transaction == nil {
		return fmt.Errorf("tx to verify not be nil")
	}

	if len(transaction.TxIn) != 1 {
		return fmt.Errorf("tx to sign must have exactly one input")
	}

	if pubKey == nil {
		return fmt.Errorf("public key must not be nil")
	}

	tapLeaf := txscript.NewBaseTapLeaf(script)

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		fundingOutputPkScript,
		fundingOutputValue,
	)

	sigHashes := txscript.NewTxSigHashes(transaction, inputFetcher)

	sigHash, err := txscript.CalcTapscriptSignaturehash(
		sigHashes, txscript.SigHashDefault, transaction, 0, inputFetcher, tapLeaf,
	)

	if err != nil {
		return err
	}

	parsedSig, err := schnorr.ParseSignature(signature)

	if err != nil {
		return err
	}

	valid := parsedSig.Verify(sigHash, pubKey)

	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}
