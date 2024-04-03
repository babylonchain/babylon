package client_test

import (
	"math/rand"
	"strings"
	"testing"

	bbn "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/client/client"
	"github.com/babylonchain/babylon/client/config"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
)

func FuzzKeys(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// create a keyring
		keyringName := datagen.GenRandomHexStr(r, 10)
		dir := t.TempDir()
		mockIn := strings.NewReader("")
		cdc := bbn.GetEncodingConfig()
		kr, err := keyring.New(keyringName, "test", dir, mockIn, cdc.Codec)
		require.NoError(t, err)

		// create a random key pair in this keyring
		keyName := datagen.GenRandomHexStr(r, 10)
		_, _, err = kr.NewMnemonic(
			keyName,
			keyring.English,
			hd.CreateHDPath(118, 0, 0).String(),
			keyring.DefaultBIP39Passphrase,
			hd.Secp256k1,
		)
		require.NoError(t, err)

		// create a Babylon client with this random keyring
		cfg := config.DefaultBabylonConfig()
		cfg.KeyDirectory = dir
		cfg.Key = keyName
		cl, err := client.New(&cfg, nil)
		require.NoError(t, err)

		// retrieve the key info from key ring
		keys, err := kr.List()
		require.NoError(t, err)
		require.Equal(t, 1, len(keys))

		// test if the key is consistent in Babylon client and keyring
		bbnAddr := cl.MustGetAddr()
		addr, _ := keys[0].GetAddress()
		require.Equal(t, addr.String(), bbnAddr)
	})
}
