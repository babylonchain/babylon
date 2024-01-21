package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func main() {
	initialSecretKey, amount, unspentTxHash, outIndex, fees := parseFlags()

	// generate keys for first tx
	fmt.Println("Generating extractable keys (address A)")
	secretKey, publicKey, privateRand, publicRand, wif1, address1 := generateKeys()
	fmt.Println("Address A: ", address1.EncodeAddress())
	fmt.Println("Address A private key (WIF):", wif1.String())

	// generate a transaction
	fmt.Println("\nGenerating first transaction spendable by extractable key...")
	txHash1, tx1, err := generateTx(initialSecretKey, amount, fees, address1, unspentTxHash, outIndex)
	if err != nil {
		panic(err)
	}
	fmt.Println("First tx hash: ", txHash1)
	fmt.Println("First tx encoded: ", tx1)

	h1, s1, h2, s2 := generateSignatures(secretKey, privateRand, "1", "2")

	// the secret key is no longer used
	secretKey = nil

	fmt.Println("---")
	fmt.Println("Extracting key...")

	extractedKey, _ := eots.Extract(publicKey, publicRand, h1, s1, h2, s2)

	fmt.Println("Key extracted")

	fmt.Println("Generating keys for address B")
	_, _, _, _, wif2, address2 := generateKeys()
	fmt.Println("Address B: ", address2.EncodeAddress())
	fmt.Println("Private key for address B: ", wif2.String())

	fmt.Println("\nGenerating transaction that spends BTC using the extracted key...")
	txHash2, tx2, err := generateTx(extractedKey, amount-fees, fees, address2, txHash1, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println("Second tx hash:", txHash2)
	fmt.Println("Second tx encoded:", tx2)
}

func generateKeys() (*eots.PrivateKey, *eots.PublicKey, *eots.PrivateRand, *eots.PublicRand, *btcutil.WIF, *btcutil.AddressPubKeyHash) {
	secretKey, err := eots.KeyGen(rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey := eots.PubGen(secretKey)

	privateRand, publicRand, err := eots.RandGen(rand.Reader)
	if err != nil {
		panic(err)
	}

	wif, err := btcutil.NewWIF(secretKey, &chaincfg.TestNet3Params, true)
	if err != nil {
		panic(err)
	}

	address, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(publicKey.SerializeCompressed()), &chaincfg.TestNet3Params)
	if err != nil {
		panic(err)
	}

	return secretKey, publicKey, privateRand, publicRand, wif, address
}

func generateSignatures(secretKey *eots.PrivateKey, privateRand *eots.PrivateRand, message1, message2 string) ([]byte, *eots.Signature, []byte, *eots.Signature) {
	h1 := chainhash.HashB([]byte(message1))
	s1, _ := eots.Sign(secretKey, privateRand, h1)

	h2 := chainhash.HashB([]byte(message2))
	s2, _ := eots.Sign(secretKey, privateRand, h2)

	return h1, s1, h2, s2
}

func generateTx(senderPk *eots.PrivateKey, amount int64, fees int64, rcvAddress *btcutil.AddressPubKeyHash, unspentTxId *chainhash.Hash, outIndex uint32) (*chainhash.Hash, string, error) {
	recTx := wire.NewMsgTx(wire.TxVersion)

	outPoint := wire.NewOutPoint(unspentTxId, outIndex)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	recTx.AddTxIn(txIn)

	rcvScript2, err := txscript.PayToAddrScript(rcvAddress)
	if err != nil {
		return nil, "", err
	}
	outCoin := int64(amount - fees)
	txOut := wire.NewTxOut(outCoin, rcvScript2)
	recTx.AddTxOut(txOut)

	senderAddress, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(senderPk.PubKey().SerializeCompressed()), &chaincfg.TestNet3Params)
	rcvScript, err := txscript.PayToAddrScript(senderAddress)
	if err != nil {
		return nil, "", err
	}

	scriptSig, err := txscript.SignatureScript(
		recTx,
		0,
		rcvScript,
		txscript.SigHashAll,
		senderPk,
		true)
	if err != nil {
		return nil, "", err
	}
	recTx.TxIn[0].SignatureScript = scriptSig

	buf := bytes.NewBuffer(make([]byte, 0, recTx.SerializeSize()))
	recTx.Serialize(buf)

	// verify transaction
	vm, err := txscript.NewEngine(rcvScript, recTx, 0, txscript.StandardVerifyFlags, nil, nil, amount, nil)
	if err != nil {
		return nil, "", err
	}
	if err := vm.Execute(); err != nil {
		return nil, "", err
	}

	hash := recTx.TxHash()
	return &hash, hex.EncodeToString(buf.Bytes()), nil
}

func parseFlags() (*eots.PrivateKey, int64, *chainhash.Hash, uint32, int64) {
	keyWif := flag.String("key", "", "Private key used for spending utxo")

	amount := flag.Int64("amount", 0, "Amount available to spend (including fees)")

	unspentTxId := flag.String("tx", "", "Hash of unspent transaction")

	outIndex := flag.Int("outIndex", 0, "Output index of unspent transaction to spend (default 0)")

	fees := flag.Int64("fees", 1000, "Amount to be left for fees in each of the two transactions (default: 1000 satoshi)")

	flag.Parse()

	if *keyWif == "" {
		fmt.Println("Please provide key (-key=WIF_ENCODED_KEY)")
		os.Exit(1)
	}
	if *amount == 0 {
		fmt.Println("Please provide amount (-amount=AMOUNT)")
		os.Exit(1)
	}
	if *unspentTxId == "" {
		fmt.Println("Please provide unspent transaction hash (-tx=HASH)")
		os.Exit(1)
	}

	wif, err := btcutil.DecodeWIF(*keyWif)

	unspentTxHash, err := chainhash.NewHashFromStr(*unspentTxId)
	if err != nil {
		panic(err)
	}

	return wif.PrivKey, *amount, unspentTxHash, uint32(*outIndex), *fees
}
