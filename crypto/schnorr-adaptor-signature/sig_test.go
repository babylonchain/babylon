package schnorr_adaptor_signature_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/require"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
)

func FuzzEncSignAndEncVerify(f *testing.F) {
	// random seeds
	f.Add([]byte("hello"))
	f.Add([]byte("1234567890!@#$%^&*()"))
	f.Add([]byte("1234567891!@#$%^&*()"))
	f.Add([]byte("1234567892!@#$%^&*()"))
	f.Add([]byte("1234567893!@#$%^&*()"))

	f.Fuzz(func(t *testing.T, msg []byte) {
		// key pair
		sk, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		pk := sk.PubKey()

		// encryption/decryption pair
		encKey, _, err := asig.GenKeyPair()
		require.NoError(t, err)

		// message hash
		msgHash := chainhash.HashB(msg)

		// encSign message
		adaptorSig, err := asig.EncSign(sk, encKey, msgHash)
		require.NoError(t, err)

		// encVerify message
		err = adaptorSig.EncVerify(pk, encKey, msgHash)
		require.NoError(t, err)
	})
}

func FuzzDecrypt(f *testing.F) {
	// random seeds
	f.Add([]byte("hello"))
	f.Add([]byte("1234567890!@#$%^&*()"))
	f.Add([]byte("1234567891!@#$%^&*()"))
	f.Add([]byte("1234567892!@#$%^&*()"))
	f.Add([]byte("1234567893!@#$%^&*()"))

	f.Fuzz(func(t *testing.T, msg []byte) {
		// key pair
		sk, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		pk := sk.PubKey()

		// encryption/decryption key pair
		encKey, decKey, err := asig.GenKeyPair()
		require.NoError(t, err)

		// message hash
		msgHash := chainhash.HashB(msg)

		// encSign message
		adaptorSig, err := asig.EncSign(sk, encKey, msgHash)
		require.NoError(t, err)

		// decrypt message
		schnorrSig := adaptorSig.Decrypt(decKey)

		// decrypted Schnorr signature should be valid
		resVerify := schnorrSig.Verify(msgHash, pk)
		require.True(t, resVerify)
	})
}

func FuzzRecover(f *testing.F) {
	// random seeds
	f.Add([]byte("hello"))
	f.Add([]byte("1234567890!@#$%^&*()"))
	f.Add([]byte("1234567891!@#$%^&*()"))
	f.Add([]byte("1234567892!@#$%^&*()"))
	f.Add([]byte("1234567893!@#$%^&*()"))

	f.Fuzz(func(t *testing.T, msg []byte) {
		// key pair
		sk, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		// encryption/decryption key pair
		encKey, decKey, err := asig.GenKeyPair()
		require.NoError(t, err)

		// message hash
		msgHash := chainhash.HashB(msg)

		// encSign message
		adaptorSig, err := asig.EncSign(sk, encKey, msgHash)
		require.NoError(t, err)

		// decrypt message
		schnorrSig := adaptorSig.Decrypt(decKey)

		// recover
		expectedDecKey := adaptorSig.Recover(schnorrSig)

		// assert the recovered decryption key is the expected one
		require.True(t, expectedDecKey.Equals(&decKey.ModNScalar))
	})
}

func FuzzSerializeAdaptorSig(f *testing.F) {
	// random seeds
	f.Add([]byte("hello"))
	f.Add([]byte("1234567890!@#$%^&*()"))
	f.Add([]byte("1234567891!@#$%^&*()"))
	f.Add([]byte("1234567892!@#$%^&*()"))
	f.Add([]byte("1234567893!@#$%^&*()"))

	f.Fuzz(func(t *testing.T, msg []byte) {
		// key pair
		sk, err := btcec.NewPrivateKey()
		require.NoError(t, err)

		// encryption/decryption key pair
		encKey, _, err := asig.GenKeyPair()
		require.NoError(t, err)

		// message hash
		msgHash := chainhash.HashB(msg)

		// encSign message
		adaptorSig, err := asig.EncSign(sk, encKey, msgHash)
		require.NoError(t, err)

		// roundtrip for serialising/deserialising adaptor signature
		adaptorSigBytes, err := adaptorSig.Marshal()
		require.NoError(t, err)
		var unmarshalledSig asig.AdaptorSignature
		err = unmarshalledSig.Unmarshal(adaptorSigBytes)
		require.NoError(t, err)
		require.True(t, adaptorSig.Equals(unmarshalledSig))

		fromBytesSig, err := asig.NewAdaptorSignatureFromBytes(adaptorSigBytes)
		require.NoError(t, err)
		require.True(t, adaptorSig.Equals(*fromBytesSig))

		sigHex := adaptorSig.MarshalHex()
		fromHexSig, err := asig.NewAdaptorSignatureFromHex(sigHex)
		require.NoError(t, err)
		require.True(t, adaptorSig.Equals(*fromHexSig))
	})
}
