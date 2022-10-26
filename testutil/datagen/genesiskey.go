package datagen

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ed255192 "github.com/tendermint/tendermint/crypto/ed25519"
)

func GenerateGenesisKey() (*types.GenesisKey, error) {
	accPrivKey := secp256k1.GenPrivKey()
	tmValPrivKey := ed255192.GenPrivKey()
	blsPrivKey := bls12381.GenPrivKey()
	tmValPubKey := tmValPrivKey.PubKey()
	valPubKey, err := codec.FromTmPubKeyInterface(tmValPubKey)
	if err != nil {
		return nil, err
	}

	blsPubKey := blsPrivKey.PubKey()
	address := sdk.ValAddress(accPrivKey.PubKey().Address())
	pop, err := privval.BuildPoP(tmValPrivKey, blsPrivKey)
	if err != nil {
		return nil, err
	}

	gk, err := types.NewGenesisKey(address, &blsPubKey, pop, valPubKey)
	if err != nil {
		return nil, err
	}

	return gk, nil
}
