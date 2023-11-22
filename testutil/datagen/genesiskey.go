package datagen

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	ed255192 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenerateGenesisKey() *types.GenesisKey {
	accPrivKey := secp256k1.GenPrivKey()
	tmValPrivKey := ed255192.GenPrivKey()
	blsPrivKey := bls12381.GenPrivKey()
	tmValPubKey := tmValPrivKey.PubKey()
	valPubKey, err := codec.FromCmtPubKeyInterface(tmValPubKey)
	if err != nil {
		panic(err)
	}

	blsPubKey := blsPrivKey.PubKey()
	address := sdk.ValAddress(accPrivKey.PubKey().Address())
	pop, err := privval.BuildPoP(tmValPrivKey, blsPrivKey)
	if err != nil {
		panic(err)
	}

	gk, err := types.NewGenesisKey(address, &blsPubKey, pop, valPubKey)
	if err != nil {
		panic(err)
	}

	return gk
}
