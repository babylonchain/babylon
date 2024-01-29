package bip322

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	bip322Tag = "BIP0322-signed-message"

	// toSpend tx constants
	toSpendVersion     = 0
	toSpendLockTime    = 0
	toSpendInputHash   = "0000000000000000000000000000000000000000000000000000000000000000"
	toSpendInputIndex  = 0xFFFFFFFF
	toSpendInputSeq    = 0
	toSpendOutputValue = 0

	// toSign tx constants
	toSignVersion     = 0
	toSignLockTime    = 0
	toSignInputSeq    = 0
	toSignOutputValue = 0
)

// GetBIP340TaggedHash builds a BIP-340 tagged hash
// More specifically, the hash is of the form
// sha256(sha256(tag) || sha256(tag) || msg)
// See https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0340.mediawiki#design
// for more details
func GetBIP340TaggedHash(msg []byte) [32]byte {
	tagHash := sha256.Sum256([]byte(bip322Tag))
	sum := make([]byte, 0)
	sum = append(sum, tagHash[:]...)
	sum = append(sum, tagHash[:]...)
	sum = append(sum, msg...)
	return sha256.Sum256(sum)
}

// toSpendSignatureScript creates the signature script for the input
// of the toSpend transaction, i.e.
// `OP_0 PUSH32 [ BIP340_TAGGED_MSG ]`
// https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0322.mediawiki#full
func toSpendSignatureScript(msg []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0)
	data := GetBIP340TaggedHash(msg)
	builder.AddData(data[:])
	script, err := builder.Script()
	if err != nil {
		// msg depends on the input, so play it safe here and don't panic
		return nil, err
	}
	return script, nil
}

// toSignPkScript creates the public key script for the output
// of the toSign transaction, i.e.
// `OP_RETURN`
// https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0322.mediawiki#full
func toSignPkScript() []byte {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	script, err := builder.Script()
	if err != nil {
		// Panic as we're building the script entirely ourselves
		panic(err)
	}
	return script
}

// GetToSpendTx builds a toSpend transaction based on the BIP-322 spec
// https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0322.mediawiki#full
// It requires as input the message that is signed and the address that produced the signature
func GetToSpendTx(msg []byte, address btcutil.Address) (*wire.MsgTx, error) {
	toSpend := wire.NewMsgTx(toSpendVersion)
	toSpend.LockTime = toSpendLockTime

	// Create a single input with dummy data based on the spec constants
	inputHash, err := chainhash.NewHashFromStr(toSpendInputHash)
	if err != nil {
		// This is a constant we have defined, so an issue here is a programming error
		panic(err)
	}
	outPoint := wire.NewOutPoint(inputHash, toSpendInputIndex)

	// The signature script containing the BIP-322 Tagged message
	script, err := toSpendSignatureScript(msg)
	if err != nil {
		return nil, err
	}
	input := wire.NewTxIn(outPoint, script, nil)
	input.Sequence = toSpendInputSeq

	// Create the output
	// The PK Script should be a pay to addr script on the provided address
	pkScript, err := txscript.PayToAddrScript(address)

	if err != nil {
		return nil, err
	}

	output := wire.NewTxOut(toSpendOutputValue, pkScript)

	toSpend.AddTxIn(input)
	toSpend.AddTxOut(output)
	return toSpend, nil
}

// GetToSignTx builds a toSign transaction based on the BIP-322 spec
// https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0322.mediawiki#full
// It requires as input the toSpend transaction that it spends
// Transaction is build without any witness, so that the witness must be filled
// by the caller.
func GetToSignTx(toSpend *wire.MsgTx) *wire.MsgTx {
	toSign := wire.NewMsgTx(toSignVersion)
	toSign.LockTime = toSignLockTime

	// Specify the input outpoint
	// Given that the input is the toSpend tx we have built, the input index is 0
	inputHash := toSpend.TxHash()
	outPoint := wire.NewOutPoint(&inputHash, 0)

	input := wire.NewTxIn(outPoint, nil, nil)
	input.Sequence = toSignInputSeq

	// Create the output
	output := wire.NewTxOut(toSignOutputValue, toSignPkScript())

	toSign.AddTxIn(input)
	toSign.AddTxOut(output)
	return toSign
}

func Verify(
	msg []byte,
	witness wire.TxWitness,
	address btcutil.Address,
	net *chaincfg.Params) error {

	toSpend, err := GetToSpendTx(msg, address)
	if err != nil {
		return err
	}

	toSign := GetToSignTx(toSpend)

	toSign.TxIn[0].Witness = witness

	// From the rules here:
	// https://github.com/bitcoin/bips/blob/master/bip-0322.mediawiki#verification-process
	// We only need to perform verification of whether toSign spends toSpend properly
	// given that the signature is a simple one and we construct both toSpend and toSign
	inputFetcher := txscript.NewCannedPrevOutputFetcher(toSpend.TxOut[0].PkScript, 0)
	sigHashes := txscript.NewTxSigHashes(toSign, inputFetcher)
	vm, err := txscript.NewEngine(
		toSpend.TxOut[0].PkScript, toSign, 0,
		txscript.StandardVerifyFlags, txscript.NewSigCache(0), sigHashes,
		toSpend.TxOut[0].Value, inputFetcher,
	)

	if err != nil {
		return err
	}

	return vm.Execute()
}

func PubkeyToP2WPKHAddress(p *btcec.PublicKey, net *chaincfg.Params) (*btcutil.AddressWitnessPubKeyHash, error) {
	witnessAddr, err := btcutil.NewAddressWitnessPubKeyHash(
		btcutil.Hash160(p.SerializeCompressed()),
		net,
	)

	if err != nil {
		return nil, err
	}

	return witnessAddr, nil
}

func PubKeyToP2TrSpendAddress(p *btcec.PublicKey, net *chaincfg.Params) (*btcutil.AddressTaproot, error) {
	tapKey := txscript.ComputeTaprootKeyNoScript(p)

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey), net,
	)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func SignWithP2WPKHAddress(
	msg []byte,
	privKey *btcec.PrivateKey,
	net *chaincfg.Params,
) (*btcutil.AddressWitnessPubKeyHash, []byte, error) {
	pubKey := privKey.PubKey()

	witnessAddr, err := PubkeyToP2WPKHAddress(pubKey, net)

	if err != nil {
		return nil, nil, err
	}

	toSpend, err := GetToSpendTx(msg, witnessAddr)

	if err != nil {
		return nil, nil, err
	}

	toSign := GetToSignTx(toSpend)

	fetcher := txscript.NewCannedPrevOutputFetcher(
		toSpend.TxOut[0].PkScript,
		toSpend.TxOut[0].Value,
	)

	hashCache := txscript.NewTxSigHashes(toSign, fetcher)

	// always use compressed pubkey
	witness, err := txscript.WitnessSignature(toSign, hashCache, 0,
		toSpend.TxOut[0].Value, toSpend.TxOut[0].PkScript, txscript.SigHashAll, privKey, true)

	if err != nil {
		return nil, nil, err
	}

	serializedWitness, err := SerializeWitness(witness)

	if err != nil {
		return nil, nil, err
	}

	return witnessAddr, serializedWitness, nil
}

func SignWithP2TrSpendAddress(
	msg []byte,
	privKey *btcec.PrivateKey,
	net *chaincfg.Params,
) (*btcutil.AddressTaproot, []byte, error) {
	pubKey := privKey.PubKey()

	witnessAddr, err := PubKeyToP2TrSpendAddress(pubKey, net)

	if err != nil {
		return nil, nil, err
	}

	toSpend, err := GetToSpendTx(msg, witnessAddr)

	if err != nil {
		return nil, nil, err
	}

	toSign := GetToSignTx(toSpend)

	fetcher := txscript.NewCannedPrevOutputFetcher(
		toSpend.TxOut[0].PkScript,
		toSpend.TxOut[0].Value,
	)

	hashCache := txscript.NewTxSigHashes(toSign, fetcher)

	witness, err := txscript.TaprootWitnessSignature(
		toSign, hashCache, 0, toSpend.TxOut[0].Value, toSpend.TxOut[0].PkScript,
		txscript.SigHashDefault, privKey,
	)

	if err != nil {
		return nil, nil, err
	}

	serializedWitness, err := SerializeWitness(witness)

	if err != nil {
		return nil, nil, err
	}

	return witnessAddr, serializedWitness, nil
}
