package types

import (
	"fmt"
	"sort"

	bbn "github.com/babylonchain/babylon/types"
)

func (v *BTCValidator) IsSlashed() bool {
	return v.SlashedBabylonHeight > 0
}

func (v *BTCValidator) ValidateBasic() error {
	// ensure fields are non-empty and well-formatted
	if v.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if v.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if _, err := v.BtcPk.ToBTCPK(); err != nil {
		return fmt.Errorf("BtcPk is not correctly formatted: %w", err)
	}
	if v.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if err := v.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// FilterTopNBTCValidators returns the top n validators based on VotingPower.
func FilterTopNBTCValidators(validators []*BTCValidatorWithMeta, n uint32) []*BTCValidatorWithMeta {
	numVals := uint32(len(validators))

	// if the given validator set is no bigger than n, no need to do anything
	if numVals <= n {
		return validators
	}

	// Sort the validators slice, from higher to lower voting power
	sort.SliceStable(validators, func(i, j int) bool {
		return validators[i].VotingPower > validators[j].VotingPower
	})

	// Return the top n elements
	return validators[:n]
}

func ExistsDup(btcPKs []bbn.BIP340PubKey) bool {
	seen := make(map[string]struct{})

	for _, btcPK := range btcPKs {
		pkStr := string(btcPK)
		if _, found := seen[pkStr]; found {
			return true
		} else {
			seen[pkStr] = struct{}{}
		}
	}

	return false
}

func NewSignatureInfo(pk *bbn.BIP340PubKey, sig *bbn.BIP340Signature) *SignatureInfo {
	return &SignatureInfo{
		Pk:  pk,
		Sig: sig,
	}
}
