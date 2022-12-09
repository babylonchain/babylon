package datagen

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenRandomValSet(n int) epochingtypes.ValidatorSet {
	power := int64(10)
	var valSet []epochingtypes.Validator
	for i := 0; i < n; i++ {
		address := sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
		val := epochingtypes.Validator{
			Addr:  address,
			Power: power,
		}
		valSet = append(valSet, val)
	}

	return epochingtypes.NewSortedValidatorSet(valSet)
}

func GenRandomPubkeysAndSigs(n int, msg []byte) ([]bls12381.PublicKey, []bls12381.Signature) {
	var blsPubkeys []bls12381.PublicKey
	var blsSigs []bls12381.Signature
	for i := 0; i < n; i++ {
		privKey := bls12381.GenPrivKey()
		pubkey := bls12381.GenPrivKey().PubKey()
		sig := bls12381.Sign(privKey, msg)
		blsPubkeys = append(blsPubkeys, pubkey)
		blsSigs = append(blsSigs, sig)
	}

	return blsPubkeys, blsSigs
}

func GenerateValidatorSetWithBLSPrivKeys(n int) (checkpointingtypes.ValidatorWithBLSSet, []bls12381.PrivateKey) {
	var (
		valSet      []*checkpointingtypes.ValidatorWithBlsKey
		blsPrivKeys []bls12381.PrivateKey
	)

	for i := 0; i < n; i++ {
		addr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
		blsPrivkey := bls12381.GenPrivKey()
		val := &checkpointingtypes.ValidatorWithBlsKey{
			ValidatorAddress: addr.String(),
			BlsPubKey:        blsPrivkey.PubKey(),
			VotingPower:      1000,
		}
		valSet = append(valSet, val)
		blsPrivKeys = append(blsPrivKeys, blsPrivkey)
	}

	return valSet, blsPrivKeys
}
