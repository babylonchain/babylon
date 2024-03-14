package btcstaking

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// length of magic prefix indentifying staking transactions
	MagicBytesLen = 4
	// 4 bytes magic bytes + 1 byte version + 32 bytes staker public key + 32 bytes finality provider public key + 2 bytes staking time
	V0OpReturnDataSize = 71

	v0OpReturnCreationErrMsg = "cannot create V0 op_return data"
)

type IdentifiableStakingInfo struct {
	StakingOutput         *wire.TxOut
	scriptHolder          *taprootScriptHolder
	timeLockPathLeafHash  chainhash.Hash
	unbondingPathLeafHash chainhash.Hash
	slashingPathLeafHash  chainhash.Hash
	OpReturnOutput        *wire.TxOut
}

func uint16ToBytes(v uint16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	return buf[:]
}

func uint16FromBytes(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, fmt.Errorf("invalid uint16 bytes length: %d", len(b))
	}

	return binary.BigEndian.Uint16(b), nil
}

// V0OpReturnData represents the data that is embedded in the OP_RETURN output
// It marshalls to exactly 71 bytes
type V0OpReturnData struct {
	MagicBytes                []byte
	Version                   byte
	StakerPublicKey           *XonlyPubKey
	FinalityProviderPublicKey *XonlyPubKey
	StakingTime               uint16
}

func NewV0OpReturnData(
	magicBytes []byte,
	stakerPublicKey []byte,
	finalityProviderPublicKey []byte,
	stakingTime []byte,
) (*V0OpReturnData, error) {
	if len(magicBytes) != MagicBytesLen {
		return nil, fmt.Errorf("%s: invalid magic bytes length: %d, expected: %d", v0OpReturnCreationErrMsg, len(magicBytes), MagicBytesLen)
	}

	stakerKey, err := XOnlyPublicKeyFromBytes(stakerPublicKey)

	if err != nil {
		return nil, fmt.Errorf("%s:invalid staker public key:%w", v0OpReturnCreationErrMsg, err)
	}

	fpKey, err := XOnlyPublicKeyFromBytes(finalityProviderPublicKey)

	if err != nil {
		return nil, fmt.Errorf("%s:invalid finality provider public key:%w", v0OpReturnCreationErrMsg, err)
	}

	stakingTimeValue, err := uint16FromBytes(stakingTime)

	if err != nil {
		return nil, fmt.Errorf("%s:invalid staking time:%w", v0OpReturnCreationErrMsg, err)
	}

	return NewV0OpReturnDataFromParsed(magicBytes, stakerKey.PubKey, fpKey.PubKey, stakingTimeValue)
}

func NewV0OpReturnDataFromParsed(
	magicBytes []byte,
	stakerPublicKey *btcec.PublicKey,
	finalityProviderPublicKey *btcec.PublicKey,
	stakingTime uint16,
) (*V0OpReturnData, error) {
	if len(magicBytes) != MagicBytesLen {
		return nil, fmt.Errorf("%s:invalid magic bytes length: %d, expected: %d", v0OpReturnCreationErrMsg, len(magicBytes), MagicBytesLen)
	}

	if stakerPublicKey == nil {
		return nil, fmt.Errorf("%s:nil staker public key", v0OpReturnCreationErrMsg)
	}

	if finalityProviderPublicKey == nil {
		return nil, fmt.Errorf("%s: nil finality provider public key", v0OpReturnCreationErrMsg)
	}

	return &V0OpReturnData{
		MagicBytes:                magicBytes,
		Version:                   0,
		StakerPublicKey:           &XonlyPubKey{stakerPublicKey},
		FinalityProviderPublicKey: &XonlyPubKey{finalityProviderPublicKey},
		StakingTime:               stakingTime,
	}, nil
}

func NewV0OpReturnDataFromBytes(b []byte) (*V0OpReturnData, error) {
	if len(b) != V0OpReturnDataSize {
		return nil, fmt.Errorf("invalid op return data length: %d, expected: %d", len(b), V0OpReturnDataSize)
	}
	magicBytes := b[:MagicBytesLen]

	version := b[MagicBytesLen]

	if version != 0 {
		return nil, fmt.Errorf("invalid op return version: %d, expected: %d", version, 0)
	}

	stakerPublicKey := b[MagicBytesLen+1 : MagicBytesLen+1+schnorr.PubKeyBytesLen]
	finalityProviderPublicKey := b[MagicBytesLen+1+schnorr.PubKeyBytesLen : MagicBytesLen+1+schnorr.PubKeyBytesLen*2]
	stakingTime := b[MagicBytesLen+1+schnorr.PubKeyBytesLen*2:]
	return NewV0OpReturnData(magicBytes, stakerPublicKey, finalityProviderPublicKey, stakingTime)
}

func getV0OpReturnBytes(out *wire.TxOut) ([]byte, error) {
	if out == nil {
		return nil, fmt.Errorf("nil tx output")
	}

	// We are adding `+2` as each op return has additional 2 for:
	// 1. OP_RETURN opcode - which signalizes that data is provably unspendable
	// 2. OP_DATA_71 opcode - which pushes 71 bytes of data to the stack
	if len(out.PkScript) != V0OpReturnDataSize+2 {
		return nil, fmt.Errorf("invalid op return data length: %d, expected: %d", len(out.PkScript), V0OpReturnDataSize+2)
	}

	if !txscript.IsNullData(out.PkScript) {
		return nil, fmt.Errorf("invalid op return script")
	}

	return out.PkScript[2:], nil
}

func NewV0OpReturnDataFromTxOutput(out *wire.TxOut) (*V0OpReturnData, error) {
	data, err := getV0OpReturnBytes(out)

	if err != nil {
		return nil, fmt.Errorf("cannot parse op return data: %w", err)
	}

	return NewV0OpReturnDataFromBytes(data)
}

func (d *V0OpReturnData) Marshall() []byte {
	var data []byte
	data = append(data, d.MagicBytes...)
	data = append(data, d.Version)
	data = append(data, d.StakerPublicKey.Marshall()...)
	data = append(data, d.FinalityProviderPublicKey.Marshall()...)
	data = append(data, uint16ToBytes(d.StakingTime)...)
	return data
}

func (d *V0OpReturnData) ToTxOutput() (*wire.TxOut, error) {
	dataScript, err := txscript.NullDataScript(d.Marshall())
	if err != nil {
		return nil, err
	}
	return wire.NewTxOut(0, dataScript), nil
}

// BuildV0IdentifiableStakingOutputs creates outputs which every staking transaction must have
func BuildV0IdentifiableStakingOutputs(
	magicBytes []byte,
	stakerKey *btcec.PublicKey,
	fpKey *btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*IdentifiableStakingInfo, error) {
	info, err := BuildStakingInfo(
		stakerKey,
		[]*btcec.PublicKey{fpKey},
		covenantKeys,
		covenantQuorum,
		stakingTime,
		stakingAmount,
		net,
	)
	if err != nil {
		return nil, err
	}

	opReturnData, err := NewV0OpReturnDataFromParsed(magicBytes, stakerKey, fpKey, stakingTime)

	if err != nil {
		return nil, err
	}

	dataOutput, err := opReturnData.ToTxOutput()

	if err != nil {
		return nil, err
	}

	return &IdentifiableStakingInfo{
		StakingOutput:         info.StakingOutput,
		scriptHolder:          info.scriptHolder,
		timeLockPathLeafHash:  info.timeLockPathLeafHash,
		unbondingPathLeafHash: info.unbondingPathLeafHash,
		slashingPathLeafHash:  info.slashingPathLeafHash,
		OpReturnOutput:        dataOutput,
	}, nil
}

// BuildV0IdentifiableStakingOutputsAndTx creates outputs which every staking transaction must have and
// returns the not-funded transaction with these outputs
func BuildV0IdentifiableStakingOutputsAndTx(
	magicBytes []byte,
	stakerKey *btcec.PublicKey,
	fpKey *btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*IdentifiableStakingInfo, *wire.MsgTx, error) {
	info, err := BuildV0IdentifiableStakingOutputs(
		magicBytes,
		stakerKey,
		fpKey,
		covenantKeys,
		covenantQuorum,
		stakingTime,
		stakingAmount,
		net,
	)
	if err != nil {
		return nil, nil, err
	}

	tx := wire.NewMsgTx(2)
	tx.AddTxOut(info.StakingOutput)
	tx.AddTxOut(info.OpReturnOutput)
	return info, tx, nil
}

func (i *IdentifiableStakingInfo) TimeLockPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.timeLockPathLeafHash)
}

func (i *IdentifiableStakingInfo) UnbondingPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.unbondingPathLeafHash)
}

func (i *IdentifiableStakingInfo) SlashingPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.slashingPathLeafHash)
}

type ParsedV0StakingTx struct {
	StakingOutput     *wire.TxOut
	StakingOutputIdx  int
	OpReturnOutput    *wire.TxOut
	OpReturnOutputIdx int
	OpReturnData      *V0OpReturnData
}

func tryToGetOpReturnDataFromOutputs(outputs []*wire.TxOut) (*V0OpReturnData, int, error) {
	// lack of outputs is not an error
	if len(outputs) == 0 {
		return nil, -1, nil
	}

	var opReturnData *V0OpReturnData
	var opReturnOutputIdx int

	for i, o := range outputs {
		output := o
		d, err := NewV0OpReturnDataFromTxOutput(output)

		if err != nil {
			// this is not an op return output recognized by Babylon, move forward
			continue
		}
		// this case should not happen as standard bitcoin node propagation rules
		// disallow multiple op return outputs in a single transaction. However, miner could
		// include multiple op return outputs in a single transaction. In such case, we should
		// return an error.
		if opReturnData != nil {
			return nil, -1, fmt.Errorf("multiple op return outputs found")
		}

		opReturnData = d
		opReturnOutputIdx = i
	}

	return opReturnData, opReturnOutputIdx, nil
}

func tryToGetStakingOutput(outputs []*wire.TxOut, stakingOutputPkScript []byte) (*wire.TxOut, int, error) {
	// lack of outputs is not an error
	if len(outputs) == 0 {
		return nil, -1, nil
	}

	var stakingOutput *wire.TxOut
	var stakingOutputIdx int

	for i, o := range outputs {
		output := o

		if !bytes.Equal(output.PkScript, stakingOutputPkScript) {
			// this is not the staking output we are looking for
			continue
		}

		if stakingOutput != nil {
			// we only allow for one staking output per transaction
			return nil, -1, fmt.Errorf("multiple staking outputs found")
		}

		stakingOutput = output
		stakingOutputIdx = i
	}

	return stakingOutput, stakingOutputIdx, nil
}

// ParseV0StakingTx takes a btc transaction and checks whether it is a staking transaction and if so parses it
// for easy data retrieval.
// It does all necessary checks to ensure that the transaction is valid staking transaction.
func ParseV0StakingTx(
	tx *wire.MsgTx,
	expectedMagicBytes []byte,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	net *chaincfg.Params,
) (*ParsedV0StakingTx, error) {
	// 1. Basic arguments checks
	if tx == nil {
		return nil, fmt.Errorf("nil tx")
	}

	if len(expectedMagicBytes) != MagicBytesLen {
		return nil, fmt.Errorf("invalid magic bytes length: %d, expected: %d", len(expectedMagicBytes), MagicBytesLen)
	}

	if len(covenantKeys) == 0 {
		return nil, fmt.Errorf("no covenant keys specified")
	}

	if covenantQuorum > uint32(len(covenantKeys)) {
		return nil, fmt.Errorf("covenant quorum is greater than the number of covenant keys")
	}

	// 2. Identify whether the transaction has expected shape
	if len(tx.TxOut) < 2 {
		return nil, fmt.Errorf("staking tx must have at least 2 outputs")
	}

	opReturnData, opReturnOutputIdx, err := tryToGetOpReturnDataFromOutputs(tx.TxOut)

	if err != nil {
		return nil, fmt.Errorf("cannot parse staking transaction: %w", err)
	}

	if opReturnData == nil {
		return nil, fmt.Errorf("transaction does not have expected op return output")
	}

	// at this point we know that transaction has op return output which seems to match
	// the expected shape. Check the magic bytes and version.
	if !bytes.Equal(opReturnData.MagicBytes, expectedMagicBytes) {
		return nil, fmt.Errorf("unexpected magic bytes: %s, expected: %s",
			hex.EncodeToString(opReturnData.MagicBytes),
			hex.EncodeToString(expectedMagicBytes),
		)
	}

	if opReturnData.Version != 0 {
		return nil, fmt.Errorf("unexpcted version: %d, expected: %d", opReturnData.Version, 0)
	}

	// 3. Op return seems to be valid V0 op return output. Now, we need to check whether
	// the staking output exists and is valid.
	stakingInfo, err := BuildStakingInfo(
		opReturnData.StakerPublicKey.PubKey,
		[]*btcec.PublicKey{opReturnData.FinalityProviderPublicKey.PubKey},
		covenantKeys,
		covenantQuorum,
		opReturnData.StakingTime,
		// we can pass 0 here, as staking amount is not used when creating taproot address
		0,
		net,
	)

	if err != nil {
		return nil, fmt.Errorf("cannot build staking info: %w", err)
	}

	stakingOutput, stakingOutputIdx, err := tryToGetStakingOutput(tx.TxOut, stakingInfo.StakingOutput.PkScript)

	if err != nil {
		return nil, fmt.Errorf("cannot parse staking transaction: %w", err)
	}

	if stakingOutput == nil {
		return nil, fmt.Errorf("staking output not found in potential staking transaction")
	}

	return &ParsedV0StakingTx{
		StakingOutput:     stakingOutput,
		StakingOutputIdx:  stakingOutputIdx,
		OpReturnOutput:    tx.TxOut[opReturnOutputIdx],
		OpReturnOutputIdx: opReturnOutputIdx,
		OpReturnData:      opReturnData,
	}, nil
}

// IsPossibleV0StakingTx checks whether transaction may be a valid staking transaction
// checks:
// 1. Whether the transaction has at least 2 outputs
// 2. have an op return output
// 3. op return output has expected magic bytes
// This function is much faster than ParseV0StakingTx, as it does not perform
// all necessary checks.
func IsPossibleV0StakingTx(tx *wire.MsgTx, expectedMagicBytes []byte) bool {
	if len(expectedMagicBytes) != MagicBytesLen {
		return false
	}

	if len(tx.TxOut) < 2 {
		return false
	}

	var possibleStakingTx = false
	for _, o := range tx.TxOut {
		output := o

		data, err := getV0OpReturnBytes(output)

		if err != nil {
			// this is not an op return output recognized by Babylon, move forward
			continue
		}

		if !bytes.Equal(data[:MagicBytesLen], expectedMagicBytes) {
			// this is not the op return output we are looking for as magic bytes do not match
			continue
		}

		if data[MagicBytesLen] != 0 {
			// this is not the v0 op return output
			continue
		}

		if possibleStakingTx {
			// this is second output that matches the magic bytes, we do not allow for multiple op return outputs
			// so this is not a valid staking transaction
			return false
		}

		possibleStakingTx = true
	}

	return possibleStakingTx
}
