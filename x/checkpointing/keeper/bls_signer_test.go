package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	val1  = epochingtypes.Validator{
		Addr:  addr1,
		Power: 10,
	}
	val2 = epochingtypes.Validator{
		Addr:  addr2,
		Power: 10,
	}
	valSet      = epochingtypes.ValidatorSet{val1, val2}
	blsPrivKey1 = bls12381.GenPrivKey()
	blsPubKey1  = blsPrivKey1.PubKey()
	blsPrivKey2 = bls12381.GenPrivKey()
	blsPubKey2  = blsPrivKey2.PubKey()
	pubkeys     = []bls12381.PublicKey{blsPubKey1, blsPubKey2}
)
