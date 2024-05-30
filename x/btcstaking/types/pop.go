package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/babylonchain/babylon/crypto/bip322"
	"github.com/babylonchain/babylon/crypto/ecdsa"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type checkStakerKey func(stakerKey *bbn.BIP340PubKey) error

type bip322Sign[A btcutil.Address] func(sg []byte,
	privKey *btcec.PrivateKey,
	net *chaincfg.Params) (A, []byte, error)

// NewPoP generates a new proof of possession that sk_Babylon and sk_BTC are held by the same person
// a proof of possession contains two signatures:
// - pop.BabylonSig = sign(sk_Babylon, pk_BTC)
// - pop.BtcSig = schnorr_sign(sk_BTC, pop.BabylonSig)
func NewPoP(babylonSK cryptotypes.PrivKey, btcSK *btcec.PrivateKey) (*ProofOfPossession, error) {
	pop := ProofOfPossession{
		BtcSigType: BTCSigType_BIP340, // by default, we use BIP-340 encoding for BTC signature
	}

	// generate pop.BabylonSig = sign(sk_Babylon, pk_BTC)
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		return nil, err
	}
	pop.BabylonSig = babylonSig

	// generate pop.BtcSig = schnorr_sign(sk_BTC, pop.BabylonSig)
	// NOTE: *schnorr.Sign has to take the hash of the message.
	// So we have to hash babylonSig before signing
	babylonSigHash := tmhash.Sum(pop.BabylonSig)
	btcSig, err := schnorr.Sign(btcSK, babylonSigHash)
	if err != nil {
		return nil, err
	}
	bip340Sig := bbn.NewBIP340SignatureFromBTCSig(btcSig)
	pop.BtcSig = bip340Sig.MustMarshal()

	return &pop, nil
}

// NewPoPBTC generates a new proof of possession that sk_BTC and the address are held by the same person
// a proof of possession contains only one signature
// - pop.BtcSig = schnorr_sign(sk_BTC, bbnAddress)
func NewPoPBTC(addr sdk.AccAddress, btcSK *btcec.PrivateKey) (*ProofOfPossessionBTC, error) {
	pop := ProofOfPossessionBTC{
		BtcSigType: BTCSigType_BIP340, // by default, we use BIP-340 encoding for BTC signature
	}

	// generate pop.BtcSig = schnorr_sign(sk_BTC, hash(bbnAddress))
	// NOTE: *schnorr.Sign has to take the hash of the message.
	// So we have to hash the address before signing
	hash := tmhash.Sum(addr.Bytes())
	btcSig, err := schnorr.Sign(btcSK, hash)
	if err != nil {
		return nil, err
	}
	bip340Sig := bbn.NewBIP340SignatureFromBTCSig(btcSig)
	pop.BtcSig = bip340Sig.MustMarshal()

	return &pop, nil
}

// NewPoPWithECDSABTCSig generates a new proof of possession where Bitcoin signature is in ECDSA format
// a proof of possession contains two signatures:
// - pop.BabylonSig = sign(sk_Babylon, pk_BTC)
// - pop.BtcSig = ecdsa_sign(sk_BTC, pop.BabylonSig)
func NewPoPWithECDSABTCSig(babylonSK cryptotypes.PrivKey, btcSK *btcec.PrivateKey) (*ProofOfPossession, error) {
	pop := ProofOfPossession{
		BtcSigType: BTCSigType_ECDSA,
	}

	// generate pop.BabylonSig = sign(sk_Babylon, pk_BTC)
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		return nil, err
	}
	pop.BabylonSig = babylonSig

	// generate pop.BtcSig = ecdsa_sign(sk_BTC, pop.BabylonSig)
	// NOTE: ecdsa.Sign has to take the message as string.
	// So we have to hex babylonSig before signing
	babylonSigHex := hex.EncodeToString(pop.BabylonSig)
	btcSig, err := ecdsa.Sign(btcSK, babylonSigHex)
	if err != nil {
		return nil, err
	}
	pop.BtcSig = btcSig

	return &pop, nil
}

func babylonSigToHexHash(babylonSig []byte) []byte {
	babylonSigHash := tmhash.Sum(babylonSig)
	babylonSigHashHex := hex.EncodeToString(babylonSigHash)
	babylonSigHashHexBytes := []byte(babylonSigHashHex)
	return babylonSigHashHexBytes
}

func newPoPWithBIP322Sig[A btcutil.Address](
	babylonSK cryptotypes.PrivKey,
	btcSK *btcec.PrivateKey,
	net *chaincfg.Params,
	bip322SignFn bip322Sign[A],
) (*ProofOfPossession, error) {
	pop := ProofOfPossession{
		BtcSigType: BTCSigType_BIP322,
	}

	// generate pop.BabylonSig = sign(sk_Babylon, pk_BTC)
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		return nil, err
	}
	pop.BabylonSig = babylonSig

	// TODO: temporary solution for MVP purposes.
	// Eventually we need to use tmhash.Sum(pop.BabylonSig) rather than bbnSigHashHexBytes
	// ref: https://github.com/babylonchain/babylon/issues/433
	bbnSigHashHexBytes := babylonSigToHexHash(pop.BabylonSig)

	address, witnessSignture, err := bip322SignFn(
		bbnSigHashHexBytes,
		btcSK,
		net,
	)

	if err != nil {
		return nil, err
	}

	bip322Sig := BIP322Sig{
		Address: address.EncodeAddress(),
		Sig:     witnessSignture,
	}

	bip322SigEncoded, err := bip322Sig.Marshal()

	if err != nil {
		return nil, err
	}
	pop.BtcSig = bip322SigEncoded

	return &pop, nil
}

func NewPoPWithBIP322P2WPKHSig(
	babylonSK cryptotypes.PrivKey,
	btcSK *btcec.PrivateKey,
	net *chaincfg.Params,
) (*ProofOfPossession, error) {
	return newPoPWithBIP322Sig(babylonSK, btcSK, net, bip322.SignWithP2WPKHAddress)
}

func NewPoPWithBIP322P2TRBIP86Sig(
	babylonSK cryptotypes.PrivKey,
	btcSK *btcec.PrivateKey,
	net *chaincfg.Params,
) (*ProofOfPossession, error) {
	return newPoPWithBIP322Sig(babylonSK, btcSK, net, bip322.SignWithP2TrSpendAddress)
}

func NewPoPFromHex(popHex string) (*ProofOfPossession, error) {
	popBytes, err := hex.DecodeString(popHex)
	if err != nil {
		return nil, err
	}
	var pop ProofOfPossession
	if err := pop.Unmarshal(popBytes); err != nil {
		return nil, err
	}
	return &pop, nil
}

func NewPoPBTCFromHex(popHex string) (*ProofOfPossessionBTC, error) {
	popBytes, err := hex.DecodeString(popHex)
	if err != nil {
		return nil, err
	}
	var pop ProofOfPossessionBTC
	if err := pop.Unmarshal(popBytes); err != nil {
		return nil, err
	}
	return &pop, nil
}

func (pop *ProofOfPossession) ToHexStr() (string, error) {
	popBytes, err := pop.Marshal()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(popBytes), nil
}

func (pop *ProofOfPossession) Verify(babylonPK cryptotypes.PubKey, bip340PK *bbn.BIP340PubKey, net *chaincfg.Params) error {
	switch pop.BtcSigType {
	case BTCSigType_BIP340:
		return pop.VerifyBIP340(babylonPK, bip340PK)
	case BTCSigType_BIP322:
		return pop.VerifyBIP322(babylonPK, bip340PK, net)
	case BTCSigType_ECDSA:
		return pop.VerifyECDSA(babylonPK, bip340PK)
	default:
		return fmt.Errorf("invalid BTC signature type")
	}
}

// Verify that the BTC private key corresponding to the bip340PK signed the staker address
func (pop *ProofOfPossessionBTC) Verify(staker sdk.AccAddress, bip340PK *bbn.BIP340PubKey, net *chaincfg.Params) error {
	switch pop.BtcSigType {
	case BTCSigType_BIP340:
		return pop.VerifyBIP340(staker, bip340PK)
	case BTCSigType_BIP322:
		return pop.VerifyBIP322(staker, bip340PK, net)
	case BTCSigType_ECDSA:
		return pop.VerifyECDSA(staker, bip340PK)
	default:
		return fmt.Errorf("invalid BTC signature type")
	}
}

// VerifyBIP340 if the BTC signature has signed the hash by the pair of bip340PK.
func VerifyBIP340(sigType BTCSigType, btcSigRaw []byte, bip340PK *bbn.BIP340PubKey, msg []byte) error {
	if sigType != BTCSigType_BIP340 {
		return fmt.Errorf("the Bitcoin signature in this proof of possession is not using BIP-340 encoding")
	}

	bip340Sig, err := bbn.NewBIP340Signature(btcSigRaw)
	if err != nil {
		return err
	}
	btcSig, err := bip340Sig.ToBTCSig()
	if err != nil {
		return err
	}
	btcPK, err := bip340PK.ToBTCPK()
	if err != nil {
		return err
	}

	// NOTE: btcSig.Verify has to take hash of the message.
	// So we have to hash babylonSig before verifying the signature
	hash := tmhash.Sum(msg)
	if !btcSig.Verify(hash, btcPK) {
		return fmt.Errorf("failed to verify pop.BtcSig")
	}

	return nil
}

// VerifyBIP340 verifies the validity of PoP where Bitcoin signature is in BIP-340
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=staker_addr)?
func (pop *ProofOfPossessionBTC) VerifyBIP340(stakerAddr sdk.AccAddress, bip340PK *bbn.BIP340PubKey) error {
	return VerifyBIP340(pop.BtcSigType, pop.BtcSig, bip340PK, stakerAddr.Bytes())
}

// VerifyBIP340 verifies the validity of PoP where Bitcoin signature is in BIP-340
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=pop.BabylonSig)?
// 2. verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
func (pop *ProofOfPossession) VerifyBIP340(babylonPK cryptotypes.PubKey, bip340PK *bbn.BIP340PubKey) error {
	if err := VerifyBIP340(pop.BtcSigType, pop.BtcSig, bip340PK, pop.BabylonSig); err != nil {
		return fmt.Errorf("failed to verify possesion of babylon sig by the BTC key: %w", err)
	}

	// rule 2: verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
	if !babylonPK.VerifySignature(*bip340PK, pop.BabylonSig) {
		return fmt.Errorf("failed to verify pop.BabylonSig")
	}

	return nil
}

// isSupportedAddressAndWitness checks whether provided address and witness are
// valid for proof of possession verification.
// Currently the only supported options are:
// 1. p2wpkh address which should only 2 elements in witness: signature and public key
// 2. p2tr address which should only 1 element in witness: signature i.e p2tr key spend
// If validation succeeds, it returns a function which can be used to check whether
// bip340PK corresponds to verified address.
func isSupportedAddressAndWitness(
	address btcutil.Address,
	witness wire.TxWitness,
	net *chaincfg.Params) (checkStakerKey, error) {
	script, err := txscript.PayToAddrScript(address)

	if err != nil {
		return nil, err
	}

	// pay to taproot key spending path have only signature in witness
	if txscript.IsPayToTaproot(script) && len(witness) == 1 {
		return func(stakerKey *bbn.BIP340PubKey) error {
			btcKey, err := stakerKey.ToBTCPK()

			if err != nil {
				return err
			}

			keyAddress, err := bip322.PubKeyToP2TrSpendAddress(btcKey, net)

			if err != nil {
				return err
			}

			if !bytes.Equal(keyAddress.ScriptAddress(), address.ScriptAddress()) {
				return fmt.Errorf("bip322Sig.Address does not correspond to bip340PK")
			}

			return nil
		}, nil
	}

	// pay to witness key hash have signature and public key in witness
	if txscript.IsPayToWitnessPubKeyHash(script) && len(witness) == 2 {
		return func(stakerKey *bbn.BIP340PubKey) error {
			keyFromWitness, err := btcec.ParsePubKey(witness[1])

			if err != nil {
				return err
			}

			keyFromWitnessBytess := schnorr.SerializePubKey(keyFromWitness)

			stakerKeyEncoded, err := stakerKey.Marshal()

			if err != nil {
				return err
			}

			if !bytes.Equal(keyFromWitnessBytess, stakerKeyEncoded) {
				return fmt.Errorf("bip322Sig.Address does not correspond to bip340PK")
			}

			return nil

		}, nil
	}

	return nil, fmt.Errorf("unsupported bip322 address type. Only supported options are p2wpkh and p2tr bip86 key spending path")
}

// VerifyBIP322SigPop verifies bip322 `signature` over `msg` and also checks whether
// `address` corresponds to `pubKeyNoCoord` in the given network.
// It supports only two type of addresses:
// 1. p2wpkh address
// 2. p2tr address which is defined in bip88
// Parameters:
// - msg: message which was signed
// - address: address which was used to sign the message
// - signature: bip322 signature over the message
// - pubKeyNoCoord: public key in 32 bytes format which was used to derive address
func VerifyBIP322SigPop(
	msg []byte,
	address string,
	signature []byte,
	pubKeyNoCoord []byte,
	net *chaincfg.Params,
) error {
	if len(msg) == 0 || len(address) == 0 || len(signature) == 0 || len(pubKeyNoCoord) == 0 {
		return fmt.Errorf("cannot verfiy bip322 signature. One of the required parameters is empty")
	}

	witness, err := bip322.SimpleSigToWitness(signature)
	if err != nil {
		return err
	}

	btcAddress, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return err
	}

	// we check whether address and witness are valid for proof of possession verification
	// before verifying bip322 signature. This is require to avoid cases in which
	// we receive some long running btc script to execute (like taproot script with 100 signatures)
	// for proof of possession, we only support two types of cases:
	// 1. address is p2wpkh address
	// 2. address is p2tr address and we are dealing with bip86 (https://github.com/bitcoin/bips/blob/master/bip-0086.mediawiki)
	// key spending path.
	// In those two cases we are able to link bip340PK public key to the btc address
	// used in bip322 signature verification.
	stakerKeyMatchesBtcAddressFn, err := isSupportedAddressAndWitness(btcAddress, witness, net)
	if err != nil {
		return err
	}

	if err := bip322.Verify(msg, witness, btcAddress, net); err != nil {
		return err
	}

	key, err := bbn.NewBIP340PubKey(pubKeyNoCoord)
	if err != nil {
		return err
	}

	// rule 3: verify bip322Sig.Address corresponds to bip340PK
	if err := stakerKeyMatchesBtcAddressFn(key); err != nil {
		return err
	}

	return nil
}

// VerifyBIP322 verifies the validity of PoP where Bitcoin signature is in BIP-322
// after decoding pop.BtcSig to bip322Sig which contains sig and address,
// 1. verify whether bip322 pop signature where msg=signedMsg
func VerifyBIP322(sigType BTCSigType, btcSigRaw []byte, bip340PK *bbn.BIP340PubKey, signedMsg []byte, net *chaincfg.Params) error {
	if sigType != BTCSigType_BIP322 {
		return fmt.Errorf("the Bitcoin signature in this proof of possession is not using BIP-322 encoding")
	}
	// unmarshal pop.BtcSig to bip322Sig
	var bip322Sig BIP322Sig
	if err := bip322Sig.Unmarshal(btcSigRaw); err != nil {
		return nil
	}

	btcKeyBytes, err := bip340PK.Marshal()
	if err != nil {
		return err
	}

	// 1. Verify Bip322 proof of possession signature
	if err := VerifyBIP322SigPop(
		signedMsg,
		bip322Sig.Address,
		bip322Sig.Sig,
		btcKeyBytes,
		net,
	); err != nil {
		return err
	}

	return nil
}

// VerifyBIP322 verifies the validity of PoP where Bitcoin signature is in BIP-322
// after decoding pop.BtcSig to bip322Sig which contains sig and address,
// 1. verify whether bip322 pop signature where msg=pop.BabylonSig
// 2. verify(sig=pop.BabylonSig, pubkey=babylonPK, msg=bip340PK)?
func (pop *ProofOfPossession) VerifyBIP322(babylonPK cryptotypes.PubKey, bip340PK *bbn.BIP340PubKey, net *chaincfg.Params) error {
	// TODO: temporary solution for MVP purposes.
	// Eventually we need to use tmhash.Sum(pop.BabylonSig) rather than bbnSigHashHexBytes
	// ref: https://github.com/babylonchain/babylon/issues/433
	bbnSigHashHexBytes := babylonSigToHexHash(pop.BabylonSig)
	if err := VerifyBIP322(pop.BtcSigType, pop.BtcSig, bip340PK, bbnSigHashHexBytes, net); err != nil {
		return fmt.Errorf("failed to verify possesion of babylon sig by the BTC key: %w", err)
	}

	// rule 2: verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
	if !babylonPK.VerifySignature(*bip340PK, pop.BabylonSig) {
		return fmt.Errorf("failed to verify pop.BabylonSig")
	}

	return nil
}

// VerifyBIP322 verifies the validity of PoP where Bitcoin signature is in BIP-322
// after decoding pop.BtcSig to bip322Sig which contains sig and address,
// 1. verify whether bip322 pop signature where msg=pop.BabylonSig
// 2. verify(sig=pop.BabylonSig, pubkey=babylonPK, msg=bip340PK)?
func (pop *ProofOfPossessionBTC) VerifyBIP322(addr sdk.AccAddress, bip340PK *bbn.BIP340PubKey, net *chaincfg.Params) error {
	return VerifyBIP322(pop.BtcSigType, pop.BtcSig, bip340PK, addr.Bytes(), net)
}

// VerifyECDSA verifies the validity of PoP where Bitcoin signature is in ECDSA encoding
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=msg)?
func VerifyECDSA(sigType BTCSigType, btcSigRaw []byte, bip340PK *bbn.BIP340PubKey, msg []byte) error {
	if sigType != BTCSigType_ECDSA {
		return fmt.Errorf("the Bitcoin signature in this proof of possession is not using ECDSA encoding")
	}

	// rule 1: verify(sig=sig_btc, pubkey=pk_btc, msg=msg)?
	btcPK, err := bip340PK.ToBTCPK()
	if err != nil {
		return err
	}
	// NOTE: ecdsa.Verify has to take message as a string
	// So we have to hex BabylonSig before verifying the signature
	bbnSigHex := hex.EncodeToString(msg)
	if err := ecdsa.Verify(btcPK, bbnSigHex, btcSigRaw); err != nil {
		return fmt.Errorf("failed to verify btcSigRaw")
	}

	return nil
}

// VerifyECDSA verifies the validity of PoP where Bitcoin signature is in ECDSA encoding
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=pop.BabylonSig)?
// 2. verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
func (pop *ProofOfPossession) VerifyECDSA(babylonPK cryptotypes.PubKey, bip340PK *bbn.BIP340PubKey) error {
	if err := VerifyECDSA(pop.BtcSigType, pop.BtcSig, bip340PK, pop.BabylonSig); err != nil {
		return fmt.Errorf("failed to verify possesion of babylon sig by the BTC key: %w", err)
	}

	// rule 2: verify(sig=pop.BabylonSig, pubkey=pk_babylon, msg=pk_btc)?
	if !babylonPK.VerifySignature(*bip340PK, pop.BabylonSig) {
		return fmt.Errorf("failed to verify pop.BabylonSig")
	}

	return nil
}

// VerifyECDSA verifies the validity of PoP where Bitcoin signature is in ECDSA encoding
// 1. verify(sig=sig_btc, pubkey=pk_btc, msg=addr)?
func (pop *ProofOfPossessionBTC) VerifyECDSA(addr sdk.AccAddress, bip340PK *bbn.BIP340PubKey) error {
	return VerifyECDSA(pop.BtcSigType, pop.BtcSig, bip340PK, addr.Bytes())
}

func (pop *ProofOfPossession) ValidateBasic() error {
	if len(pop.BabylonSig) == 0 {
		return fmt.Errorf("empty Babylon signature")
	}
	if pop.BtcSig == nil {
		return fmt.Errorf("empty BTC signature")
	}

	return nil
}

// ValidateBasic checks if there is a BTC Signature.
func (pop *ProofOfPossessionBTC) ValidateBasic() error {
	if pop.BtcSig == nil {
		return fmt.Errorf("empty BTC signature")
	}

	return nil
}
