package eots_test

import (
	"bytes"
	"crypto/rand"
	mathrand "math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
	"github.com/vulpine-io/io-test/v1/pkg/iotest"
)

// TODO: possible improvements
// test KeyGen, PubGen, RandGen give consistent results with deterministic randomness source
// test compare signatures against btcec

func FuzzSignAndVerify(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := mathrand.New(mathrand.NewSource(seed))

		sk, err := eots.KeyGen(r)
		require.NoError(t, err)
		pk := eots.PubGen(sk)

		sr, pr, err := eots.RandGen(rand.Reader)
		require.NoError(t, err)

		msg := datagen.GenRandomByteArray(r, 100)
		sig, err := eots.Sign(sk, sr, msg)
		require.NoError(t, err)

		err = eots.Verify(pk, pr, msg, sig)
		require.NoError(t, err)
	})
}

func TestSignAndInvalidVerify(t *testing.T) {
	randSource := new(iotest.ReadCloser)
	sk, err := eots.KeyGen(randSource)
	if err != nil {
		t.Fatal(err)
	}
	k, publicK, err := eots.RandGen(randSource)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("hello")

	sig, err := eots.Sign(sk, k, message)
	if err != nil {
		t.Fatal(err)
	}

	invalidK := new(secp256k1.FieldVal).Set(publicK).AddInt(1)

	err = eots.Verify(eots.PubGen(sk), invalidK, message, sig)
	if err == nil {
		t.Fatal("Expected verify to fail with wrong k value")
	}

	err = eots.Verify(eots.PubGen(sk), publicK, message, new(eots.Signature))
	if err == nil {
		t.Fatal("Expected verify to fail with wrong signature for the hash")
	}

	messageInvalid := []byte("bye")
	err = eots.Verify(eots.PubGen(sk), publicK, messageInvalid, sig)
	if err == nil {
		t.Fatal("Expected verify to fail with wrong signature for the hash")
	}
}

func FuzzExtract(f *testing.F) {
	randSource := new(iotest.ReadCloser)
	sk, err := eots.KeyGen(randSource)
	if err != nil {
		f.Fatal(err)
	}
	k, publicK, err := eots.RandGen(rand.Reader)
	if err != nil {
		f.Fatal(err)
	}

	type tc struct {
		m1 []byte
		m2 []byte
	}

	for _, seed := range []tc{{[]byte("hello"), []byte("bye")}, {[]byte("1234567890"), []byte("!@#$%^&*()")}} {
		f.Add(seed.m1, seed.m2)
	}

	f.Fuzz(func(t *testing.T, message1, message2 []byte) {
		if bytes.Equal(message1, message2) {
			t.Skip()
		}

		sig1, err := eots.Sign(sk, k, message1)
		if err != nil {
			t.Fatal(err)
		}

		sig2, err := eots.Sign(sk, k, message2)
		if err != nil {
			t.Fatal(err)
		}

		sk2, err := eots.Extract(sk.PubKey(), publicK, message1, sig1, message2, sig2)
		if err != nil {
			t.Fatal(err)
		}

		if !sk.Key.Equals(&sk2.Key) {
			t.Fatal("Unexpected extracted private key")
		}
	})
}
