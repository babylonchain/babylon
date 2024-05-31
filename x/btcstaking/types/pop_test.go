package types_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg"

	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
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
func FuzzPop_ValidBip322SigNotMatchingBip340PubKey(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

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
	})
}

func TestPoPBTCValidateBasic(t *testing.T) {
	r := rand.New(rand.NewSource(10))

	btcSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	addrToSign := sdk.MustAccAddressFromBech32(datagen.GenRandomAccount().Address)
	sigHash := tmhash.Sum(addrToSign)
	btcSig, err := schnorr.Sign(btcSK, sigHash)
	require.NoError(t, err)

	tcs := []struct {
		title  string
		pop    types.ProofOfPossessionBTC
		expErr error
	}{
		{
			"valid: some sig",
			types.ProofOfPossessionBTC{
				BtcSig: []byte("something"),
			},
			nil,
		},
		{
			"valid: correct signature",
			types.ProofOfPossessionBTC{
				BtcSig: btcSig.Serialize(),
			},
			nil,
		},
		{
			"invalid: nil sig",
			types.ProofOfPossessionBTC{},
			fmt.Errorf("empty BTC signature"),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			actErr := tc.pop.ValidateBasic()
			if tc.expErr != nil {
				require.EqualError(t, actErr, tc.expErr.Error())
				return
			}
			require.NoError(t, actErr)
		})
	}
}
