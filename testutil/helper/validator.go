package helper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	cmtsecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
)

type testValidator struct {
	consAddr   sdk.ConsAddress
	tmPk       cmtprotocrypto.PublicKey
	valPrivKey cmtsecp256k1.PrivKey
	blsPrivKey bls12381.PrivateKey
}

func newTestValidator(valPrivKey cmtsecp256k1.PrivKey, blsPrivKey bls12381.PrivateKey) testValidator {
	pubkey := valPrivKey.PubKey()
	tmPk := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Secp256K1{
			Secp256K1: pubkey.Bytes(),
		},
	}

	return testValidator{
		consAddr:   sdk.ConsAddress(pubkey.Address()),
		tmPk:       tmPk,
		valPrivKey: valPrivKey,
		blsPrivKey: blsPrivKey,
	}
}

func (t testValidator) toValidator(power int64) abci.Validator {
	return abci.Validator{
		Address: t.consAddr.Bytes(),
		Power:   power,
	}
}
