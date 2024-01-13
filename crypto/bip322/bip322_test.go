package bip322_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/babylonchain/babylon/crypto/bip322"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

var (
	net                   = &chaincfg.TestNet3Params
	emptyBytes            = []byte("")
	helloWorldBytes       = []byte("Hello World")
	emptyBytesSig, _      = base64.StdEncoding.DecodeString("AkcwRAIgM2gBAQqvZX15ZiysmKmQpDrG83avLIT492QBzLnQIxYCIBaTpOaD20qRlEylyxFSeEA2ba9YOixpX8z46TSDtS40ASECx/EgAxlkQpQ9hYjgGu6EBCPMVPwVIVJqO4XCsMvViHI=")
	helloWorldBytesSig, _ = base64.StdEncoding.DecodeString("AkcwRAIgZRfIY3p7/DoVTty6YZbWS71bc5Vct9p9Fia83eRmw2QCICK/ENGfwLtptFluMGs2KsqoNSk89pO7F29zJLUx9a/sASECx/EgAxlkQpQ9hYjgGu6EBCPMVPwVIVJqO4XCsMvViHI=")
	testAddr              = "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l"
)

// test vectors at https://github.com/bitcoin/bips/blob/master/bip-0322.mediawiki#message-hashing
func TestBIP322_MsgHash(t *testing.T) {
	msgHash := bip322.GetBIP340TaggedHash(emptyBytes)
	msgHashHex := hex.EncodeToString(msgHash[:])
	require.Equal(t, msgHashHex, "c90c269c4f8fcbe6880f72a721ddfbf1914268a794cbb21cfafee13770ae19f1")

	msgHash = bip322.GetBIP340TaggedHash(helloWorldBytes)
	msgHashHex = hex.EncodeToString(msgHash[:])
	require.Equal(t, msgHashHex, "f0eb03b1a75ac6d9847f55c624a99169b5dccba2a31f5b23bea77ba270de0a7a")
}

// test vectors at https://github.com/bitcoin/bips/blob/master/bip-0322.mediawiki#transaction-hashes
func TestBIP322_TxHashToSpend(t *testing.T) {
	// empty str
	toSpendTx, err := bip322.GetToSpendTx(emptyBytes, testAddr, net)
	require.NoError(t, err)
	require.Equal(t, "c5680aa69bb8d860bf82d4e9cd3504b55dde018de765a91bb566283c545a99a7", toSpendTx.TxHash().String())
	toSignTx, err := bip322.GetToSignTx(toSpendTx, emptyBytesSig)
	require.NoError(t, err)
	require.Equal(t, "1e9654e951a5ba44c8604c4de6c67fd78a27e81dcadcfe1edf638ba3aaebaed6", toSignTx.TxHash().String())

	// hello world str
	toSpendTx, err = bip322.GetToSpendTx(helloWorldBytes, testAddr, net)
	require.NoError(t, err)
	require.Equal(t, "b79d196740ad5217771c1098fc4a4b51e0535c32236c71f1ea4d61a2d603352b", toSpendTx.TxHash().String())
	toSignTx, err = bip322.GetToSignTx(toSpendTx, helloWorldBytesSig)
	require.NoError(t, err)
	require.Equal(t, "88737ae86f2077145f93cc4b153ae9a1cb8d56afa511988c149c5c8c9d93bddf", toSignTx.TxHash().String())
}

func TestBIP322_Verify(t *testing.T) {
	sigBase64 := "AkcwRAIgbAFRpM0rhdBlXr7qe5eEf3XgSeausCm2XTmZVxSYpcsCIDcbR87wF9DTrvdw1czYEEzOjso52dOSaw8VrC4GgzFRASECO5NGNFlPClJnTHNDW94h7pPL5D7xbl6FBNTrGaYpYcA="
	msgBase64 := "HRQD77+9dmnvv71N77+9O2/Wuzbvv73vv71a77+977+977+977+9Du+/ve+/vTgrNH/vv71lQX0="
	// TODO: make it work with the public key??
	address := "tb1qfwtfzdagj7efph6zfcv68ce3v48c8e9fatunur"
	emptyBytesSig, err := base64.StdEncoding.DecodeString(sigBase64)
	require.NoError(t, err)

	msg, err := base64.StdEncoding.DecodeString(msgBase64)
	require.NoError(t, err)

	err = bip322.Verify(msg, emptyBytesSig, address, net)
	require.NoError(t, err)
}

// TODO: test signature generation
