package datagen

import (
	"fmt"
	"math/rand"

	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func GenRandomBTCValidator(r *rand.Rand) (*bstypes.BTCValidator, error) {
	// key pairs
	btcSK, btcPK, err := GenRandomBTCKeyPair(r)
	if err != nil {
		return nil, err
	}
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	if err != nil {
		return nil, err
	}
	bbnSK, bbnPK, err := GenRandomSecp256k1KeyPair(r)
	if err != nil {
		return nil, err
	}
	secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
	if !ok {
		return nil, fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
	}
	// pop
	pop, err := bstypes.NewPoP(bbnSK, btcSK)
	if err != nil {
		return nil, err
	}
	return &bstypes.BTCValidator{
		BabylonPk: secp256k1PK,
		BtcPk:     &bip340PK,
		Pop:       pop,
	}, nil
}
