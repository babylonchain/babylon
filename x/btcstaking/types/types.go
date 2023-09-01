package types

import (
	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
)

type PublicKeyInfo struct {
	StakerKey    *bbn.BIP340PubKey
	ValidatorKey *bbn.BIP340PubKey
	JuryKey      *bbn.BIP340PubKey
}

func KeyDataFromScript(scriptData *btcstaking.StakingScriptData) *PublicKeyInfo {
	return &PublicKeyInfo{
		StakerKey:    bbn.NewBIP340PubKeyFromBTCPK(scriptData.StakerKey),
		ValidatorKey: bbn.NewBIP340PubKeyFromBTCPK(scriptData.ValidatorKey),
		JuryKey:      bbn.NewBIP340PubKeyFromBTCPK(scriptData.JuryKey),
	}
}
