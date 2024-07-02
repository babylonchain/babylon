package types_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/btcsuite/btcd/chaincfg"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

var (
	net = &chaincfg.TestNet3Params
)

func newInvalidBIP340PoP(r *rand.Rand) *types.ProofOfPossessionBTC {
	return &types.ProofOfPossessionBTC{
		BtcSigType: types.BTCSigType_BIP340,
		BtcSig:     datagen.GenRandomByteArray(r, 32), // fake sig hash
	}
}

func FuzzPoP_BIP340(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		accAddr := datagen.GenRandomAccount().GetAddress()

		// generate and verify PoP, correct case
		pop, err := types.NewPoPBTC(accAddr, btcSK)
		require.NoError(t, err)
		err = pop.VerifyBIP340(accAddr, bip340PK)
		require.NoError(t, err)

		// generate and verify PoP, invalid case
		invalidPoP := newInvalidBIP340PoP(r)
		err = invalidPoP.VerifyBIP340(accAddr, bip340PK)
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

		accAddr := datagen.GenRandomAccount().GetAddress()

		// generate and verify PoP, correct case
		pop, err := types.NewPoPBTCWithECDSABTCSig(accAddr, btcSK)
		require.NoError(t, err)
		err = pop.VerifyECDSA(accAddr, bip340PK)
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

		accAddr := datagen.GenRandomAccount().GetAddress()

		// generate and verify PoP, correct case
		pop, err := types.NewPoPBTCWithBIP322P2WPKHSig(accAddr, btcSK, net)
		require.NoError(t, err)
		err = pop.VerifyBIP322(accAddr, bip340PK, net)
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

		accAddr := datagen.GenRandomAccount().GetAddress()

		// generate and verify PoP, correct case
		pop, err := types.NewPoPBTCWithBIP322P2TRBIP86Sig(accAddr, btcSK, net)
		require.NoError(t, err)
		err = pop.VerifyBIP322(accAddr, bip340PK, net)
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

		accAddr := datagen.GenRandomAccount().GetAddress()

		// generate valid bip322 P2WPKH pop
		pop, err := types.NewPoPBTCWithBIP322P2WPKHSig(accAddr, btcSK, net)
		require.NoError(t, err)

		// verify bip322 pop with incorrect staker key
		err = pop.VerifyBIP322(accAddr, bip340PK1, net)
		require.Error(t, err)

		// generate valid bip322 P2Tr pop
		pop, err = types.NewPoPBTCWithBIP322P2TRBIP86Sig(accAddr, btcSK, net)
		require.NoError(t, err)

		// verify bip322 pop with incorrect staker key
		err = pop.VerifyBIP322(accAddr, bip340PK1, net)
		require.Error(t, err)
	})
}

func TestPoPBTCValidateBasic(t *testing.T) {
	r := rand.New(rand.NewSource(10))

	btcSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	addrToSign := sdk.MustAccAddressFromBech32(datagen.GenRandomAccount().Address)

	popBip340, err := types.NewPoPBTC(addrToSign, btcSK)
	require.NoError(t, err)

	popBip322, err := types.NewPoPBTCWithBIP322P2WPKHSig(addrToSign, btcSK, &chaincfg.MainNetParams)
	require.NoError(t, err)

	popECDSA, err := types.NewPoPBTCWithECDSABTCSig(addrToSign, btcSK)
	require.NoError(t, err)

	tcs := []struct {
		title  string
		pop    *types.ProofOfPossessionBTC
		expErr error
	}{
		{
			"valid: BIP 340",
			popBip340,
			nil,
		},
		{
			"valid: BIP 322",
			popBip322,
			nil,
		},
		{
			"valid: ECDSA",
			popECDSA,
			nil,
		},
		{
			"invalid: nil sig",
			&types.ProofOfPossessionBTC{},
			fmt.Errorf("empty BTC signature"),
		},
		{
			"invalid: BIP 340 - bad sig",
			&types.ProofOfPossessionBTC{
				BtcSigType: types.BTCSigType_BIP340,
				BtcSig:     popBip322.BtcSig,
			},
			fmt.Errorf("invalid BTC BIP340 signature: bytes cannot be converted to a *schnorr.Signature object"),
		},
		{
			"invalid: BIP 322 - bad sig",
			&types.ProofOfPossessionBTC{
				BtcSigType: types.BTCSigType_BIP322,
				BtcSig:     []byte("ss"),
			},
			fmt.Errorf("invalid BTC BIP322 signature: unexpected EOF"),
		},
		{
			"invalid: ECDSA - bad sig",
			&types.ProofOfPossessionBTC{
				BtcSigType: types.BTCSigType_ECDSA,
				BtcSig:     popBip340.BtcSig,
			},
			fmt.Errorf("invalid BTC ECDSA signature size"),
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

func TestPoPBTCVerify(t *testing.T) {
	r := rand.New(rand.NewSource(10))

	addrToSign := sdk.MustAccAddressFromBech32(datagen.GenRandomAccount().Address)
	randomAddr := sdk.MustAccAddressFromBech32(datagen.GenRandomAccount().Address)

	// generate BTC key pair
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

	netParams := &chaincfg.MainNetParams

	popBip340, err := types.NewPoPBTC(addrToSign, btcSK)
	require.NoError(t, err)

	popBip322, err := types.NewPoPBTCWithBIP322P2WPKHSig(addrToSign, btcSK, netParams)
	require.NoError(t, err)

	popECDSA, err := types.NewPoPBTCWithECDSABTCSig(addrToSign, btcSK)
	require.NoError(t, err)

	tcs := []struct {
		title  string
		staker sdk.AccAddress
		btcPK  *bbn.BIP340PubKey
		pop    *types.ProofOfPossessionBTC
		expErr error
	}{
		{
			"valid: BIP340",
			addrToSign,
			bip340PK,
			popBip340,
			nil,
		},
		{
			"valid: BIP322",
			addrToSign,
			bip340PK,
			popBip322,
			nil,
		},
		{
			"valid: ECDSA",
			addrToSign,
			bip340PK,
			popECDSA,
			nil,
		},
		{
			"invalid: BIP340 - bad addr",
			randomAddr,
			bip340PK,
			popBip340,
			fmt.Errorf("failed to verify pop.BtcSig"),
		},
		{
			"invalid: BIP322 - bad addr",
			randomAddr,
			bip340PK,
			popBip322,
			fmt.Errorf("failed to verify possession of babylon sig by the BTC key: signature not empty on failed checksig"),
		},
		{
			"invalid: ECDSA - bad addr",
			randomAddr,
			bip340PK,
			popECDSA,
			fmt.Errorf("failed to verify btcSigRaw"),
		},
		{
			"invalid: SigType",
			nil,
			nil,
			&types.ProofOfPossessionBTC{
				BtcSigType: types.BTCSigType(123),
			},
			fmt.Errorf("invalid BTC signature type"),
		},
		{
			"invalid: nil sig",
			randomAddr,
			bip340PK,
			&types.ProofOfPossessionBTC{
				BtcSigType: types.BTCSigType_BIP322,
				BtcSig:     nil,
			},
			fmt.Errorf("failed to verify possession of babylon sig by the BTC key: cannot verfiy bip322 signature. One of the required parameters is empty"),
		},
		{
			"invalid: nil signed msg",
			nil,
			bip340PK,
			popBip340,
			fmt.Errorf("failed to verify pop.BtcSig"),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			actErr := tc.pop.Verify(tc.staker, tc.btcPK, netParams)
			if tc.expErr != nil {
				require.EqualError(t, actErr, tc.expErr.Error())
				return
			}
			require.NoError(t, actErr)
		})
	}
}
