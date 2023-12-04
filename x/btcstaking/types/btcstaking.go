package types

import (
	"fmt"
	"sort"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
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

// GetOrderedCovenantSignatures returns the ordered covenant adaptor signatures
// encrypted by the BTC validator's PK at the given index from the given list of
// covenant signatures
// the order of covenant adaptor signatures will follow the reverse lexicographical order
// of signing public keys, in order to be used as tx witness
func GetOrderedCovenantSignatures(valIdx int, covSigsList []*CovenantAdaptorSignatures, params *Params) ([]*asig.AdaptorSignature, error) {
	// construct the map where key is the covenant PK and value is this
	// covenant member's adaptor signature encrypted by the given validator's PK
	covSigsMap := map[string]*asig.AdaptorSignature{}
	for _, covSigs := range covSigsList {
		// find the adaptor signature at the corresponding BTC validator's index
		if valIdx >= len(covSigs.AdaptorSigs) {
			return nil, fmt.Errorf("validator index is out of the scope")
		}
		covSigBytes := covSigs.AdaptorSigs[valIdx]
		// decode the adaptor signature bytes
		covSig, err := asig.NewAdaptorSignatureFromBytes(covSigBytes)
		if err != nil {
			return nil, err
		}
		// append to map
		covSigsMap[covSigs.CovPk.MarshalHex()] = covSig
	}

	// sort covenant PKs in reverse reverse lexicographical order
	orderedCovenantPKs := bbn.SortBIP340PKs(params.CovenantPks)

	// get ordered list of covenant signatures w.r.t. the order of sorted covenant PKs
	// Note that only a quorum number of covenant signatures needs to be provided
	orderedCovSigs := []*asig.AdaptorSignature{}
	for _, covPK := range orderedCovenantPKs {
		if covSig, ok := covSigsMap[covPK.MarshalHex()]; ok {
			orderedCovSigs = append(orderedCovSigs, covSig)
		} else {
			orderedCovSigs = append(orderedCovSigs, nil)
		}
	}

	return orderedCovSigs, nil
}
