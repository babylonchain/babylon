package keeper

import (
	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

type BlsSigner interface {
	GetAddress() sdk.ValAddress
	SignMsgWithBls(msg []byte) (bls12381.Signature, error)
	GetBlsPubkey() (bls12381.PublicKey, error)
	GetValidatorPubkey() (crypto.PubKey, error)
}

// SignBLS signs a BLS signature over the given information
func (k Keeper) SignBLS(epochNum uint64, blockHash types.BlockHash) (bls12381.Signature, error) {
	// get BLS signature by signing
	signBytes := types.GetSignBytes(epochNum, blockHash)
	return k.blsSigner.SignMsgWithBls(signBytes)
}

func (k Keeper) GetBLSSignerAddress() sdk.ValAddress {
	return k.blsSigner.GetAddress()
}

func (k Keeper) GetValidatorAddress() sdk.ValAddress {
	pk, err := k.blsSigner.GetValidatorPubkey()
	if err != nil {
		panic(err)
	}
	return sdk.ValAddress(pk.Address())
}
