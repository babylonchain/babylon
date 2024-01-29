package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"
)

var (
	net = &chaincfg.TestNet3Params
)

func newInvalidBIP340PoP(r *rand.Rand, babylonSK cryptotypes.PrivKey, btcSK *btcec.PrivateKey) *types.ProofOfPossession {
	pop := types.ProofOfPossession{}

	randomNum := datagen.RandomInt(r, 2) // 0 or 1

	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		panic(err)
	}

	var babylonSigHash []byte
	if randomNum == 0 {
		pop.BabylonSig = babylonSig                        // correct sig
		babylonSigHash = datagen.GenRandomByteArray(r, 32) // fake sig hash
	} else {
		pop.BabylonSig = datagen.GenRandomByteArray(r, uint64(len(babylonSig))) // fake sig
		babylonSigHash = tmhash.Sum(pop.BabylonSig)                             // correct sig hash
	}

	btcSig, err := schnorr.Sign(btcSK, babylonSigHash)
	if err != nil {
		panic(err)
	}
	bip340Sig := bbn.NewBIP340SignatureFromBTCSig(btcSig)
	pop.BtcSig = bip340Sig.MustMarshal()

	return &pop
}

func FuzzPoP_BIP340(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		// generate Babylon key pair
		babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
		require.NoError(t, err)

		// generate and verify PoP, correct case
		pop, err := types.NewPoP(babylonSK, btcSK)
		require.NoError(t, err)
		err = pop.VerifyBIP340(babylonPK, bip340PK)
		require.NoError(t, err)

		// generate and verify PoP, invalid case
		invalidPoP := newInvalidBIP340PoP(r, babylonSK, btcSK)
		err = invalidPoP.VerifyBIP340(babylonPK, bip340PK)
		require.Error(t, err)
	})
}

func FuzzPoP_ECDSA(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		// generate Babylon key pair
		babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
		require.NoError(t, err)

		// generate and verify PoP, correct case
		pop, err := types.NewPoPWithECDSABTCSig(babylonSK, btcSK)
		require.NoError(t, err)
		err = pop.VerifyECDSA(babylonPK, bip340PK)
		require.NoError(t, err)
	})
}

func FuzzPoP_BIP322_P2WPKH(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		// generate Babylon key pair
		babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
		require.NoError(t, err)

		// generate and verify PoP, correct case
		pop, err := types.NewPoPWithBIP322P2WPKHSig(babylonSK, btcSK, net)
		require.NoError(t, err)
		err = pop.VerifyBIP322(babylonPK, bip340PK, net)
		require.NoError(t, err)
	})
}

func FuzzPoP_BIP322_P2Tr_BIP86(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		// generate Babylon key pair
		babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
		require.NoError(t, err)

		// generate and verify PoP, correct case
		pop, err := types.NewPoPWithBIP322P2TRBIP86Sig(babylonSK, btcSK, net)
		require.NoError(t, err)
		err = pop.VerifyBIP322(babylonPK, bip340PK, net)
		require.NoError(t, err)
	})
}

// TODO: Add more negative cases
func TestValidBip322SigNotMatchingBip340PubKey(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	// generate two BTC key pairs
	btcSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	_, btcPK1, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	bip340PK1 := bbn.NewBIP340PubKeyFromBTCPK(btcPK1)

	// generate Babylon key pair
	babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
	require.NoError(t, err)

	// generate valid bip322 P2WPKH pop
	pop, err := types.NewPoPWithBIP322P2WPKHSig(babylonSK, btcSK, net)
	require.NoError(t, err)

	// verify bip322 pop with incorrect staker key
	err = pop.VerifyBIP322(babylonPK, bip340PK1, net)
	require.Error(t, err)

	// generate valid bip322 P2Tr pop
	pop, err = types.NewPoPWithBIP322P2TRBIP86Sig(babylonSK, btcSK, net)
	require.NoError(t, err)

	// verify bip322 pop with incorrect staker key
	err = pop.VerifyBIP322(babylonPK, bip340PK1, net)
	require.Error(t, err)
}
