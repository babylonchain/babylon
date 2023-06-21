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

//End of copied methods

// StakingScriptData is a struct that holds data parsed from staking script
type StakingScriptData struct {
	StakerKey    *btcec.PublicKey
	ValidatorKey *btcec.PublicKey
	JuryKey      *btcec.PublicKey
	StakingTime  uint16
}

func NewStakingScriptData(
	stakerKey,
	validatorKey,
	juryKey *btcec.PublicKey,
	stakingTime uint16) (*StakingScriptData, error) {

	if stakerKey == nil || validatorKey == nil || juryKey == nil {
		return nil, fmt.Errorf("staker, validator and jury keys cannot be nil")
	}

	return &StakingScriptData{
		StakerKey:    stakerKey,
		ValidatorKey: validatorKey,
		JuryKey:      juryKey,
		StakingTime:  stakingTime,
	}, nil
}

// BuildStakingScript builds a staking script in the following format:
// <StakerKey> OP_CHECKSIG
// OP_NOTIF
//
//	<StakerKey> OP_CHECKSIG <ValidatorKey> OP_CHECKSIGADD <JuryKey> OP_CHECKSIGADD 3 OP_NUMEQUAL
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
	builder.AddData(schnorr.SerializePubKey(sd.JuryKey))
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
func ParseStakingTransactionScript(version uint16, script []byte) (*StakingScriptData, error) {
	// <StakerKey> OP_CHECKSIG
	// OP_NOTIF
	//
	//	<StakerKey> OP_CHECKSIG <ValidatorKey> OP_CHECKSIGADD <JuryKey> OP_CHECKSIGADD 3 OP_NUMEQUAL
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
	tokenizer := txscript.MakeScriptTokenizer(version, script)
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

	// Jury public key
	juryPk, err := schnorr.ParsePubKey(template[7].extractedData)

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
	scriptData, _ := NewStakingScriptData(stakerPk1, validatorPk, juryPk, uint16(template[12].extractedInt))

	return scriptData, nil
}

func UnspendableKeyPathInternalPubKey() btcec.PublicKey {
	// TODO: We could cache it in some cached private package variable if performance
	// is necessary, as this returns always the same value.
	keyBytes, _ := hex.DecodeString(unspendableKeyPath)
	// We are using btcec here, as key is 33 byte compressed format.
	pubKey, _ := btcec.ParsePubKey(keyBytes)
	return *pubKey
}

// TaprootAddressForScript returns a Taproot address commiting to the given pkScript
func TaprootAddressForScript(
	pkScript []byte,
	internalPubKey *btcec.PublicKey,
	net *chaincfg.Params) (*btcutil.AddressTaproot, error) {

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)

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

// BuildStakingOutput builds out which is necessary for staking transaction to stake funds.
func BuildStakingOutput(
	stakerKey,
	validatorKey,
	juryKey *btcec.PublicKey,
	stTime uint16,
	stAmount btcutil.Amount,
	net *chaincfg.Params) (*wire.TxOut, []byte, error) {

	sd, err := NewStakingScriptData(stakerKey, validatorKey, juryKey, stTime)

	if err != nil {
		return nil, nil, err
	}

	script, err := sd.BuildStakingScript()

	if err != nil {
		return nil, nil, err
	}

	internalPubKey := UnspendableKeyPathInternalPubKey()

	address, err := TaprootAddressForScript(script, &internalPubKey, net)

	if err != nil {
		return nil, nil, err
	}

	pkScript, err := txscript.PayToAddrScript(address)

	if err != nil {
		return nil, nil, err
	}

	return wire.NewTxOut(int64(stAmount), pkScript), script, nil
}

// BuildWitnessToSpendStakingOutput builds witness for spending staking as single staker
// Current assumptions:
// - staking output is the only input to the transaction
// - staking output is valid staking output
func BuildWitnessToSpendStakingOutput(
	tx *wire.MsgTx,
	stakingOutput *wire.TxOut,
	stakingScript []byte,
	privKey *btcec.PrivateKey) (wire.TxWitness, error) {

	internalPubKey := UnspendableKeyPathInternalPubKey()

	tapLeaf := txscript.NewBaseTapLeaf(stakingScript)

	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
		&internalPubKey,
	)

	ctrlBlockBytes, err := ctrlBlock.ToBytes()

	if err != nil {
		return nil, err
	}

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingOutput.PkScript,
		stakingOutput.Value,
	)

	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		tx, sigHashes, 0, stakingOutput.Value,
		stakingOutput.PkScript, tapLeaf, txscript.SigHashDefault,
		privKey,
	)

	if err != nil {
		return nil, err
	}

	witnessStack := wire.TxWitness(make([][]byte, 3))
	witnessStack[0] = sig
	witnessStack[1] = stakingScript
	witnessStack[2] = ctrlBlockBytes
	return witnessStack, nil
}
