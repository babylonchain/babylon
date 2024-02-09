package types

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (fp *FinalityProvider) IsSlashed() bool {
	return fp.SlashedBabylonHeight > 0
}

func (fp *FinalityProvider) ValidateBasic() error {
	// ensure fields are non-empty and well-formatted
	if fp.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if fp.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if _, err := fp.BtcPk.ToBTCPK(); err != nil {
		return fmt.Errorf("BtcPk is not correctly formatted: %w", err)
	}
	if fp.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if err := fp.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// FilterTopNFinalityProviders returns the top n finality providers based on VotingPower.
func FilterTopNFinalityProviders(fps []*FinalityProviderDistInfo, n uint32) []*FinalityProviderDistInfo {
	numFps := uint32(len(fps))

	// if the given finality provider set is no bigger than n, no need to do anything
	if numFps <= n {
		return fps
	}

	// Sort the finality providers slice, from higher to lower voting power
	sort.SliceStable(fps, func(i, j int) bool {
		return fps[i].TotalVotingPower > fps[j].TotalVotingPower
	})

	// Return the top n elements
	return fps[:n]
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
// encrypted by the finality provider's PK at the given index from the given list of
// covenant signatures
// the order of covenant adaptor signatures will follow the reverse lexicographical order
// of signing public keys, in order to be used as tx witness
func GetOrderedCovenantSignatures(fpIdx int, covSigsList []*CovenantAdaptorSignatures, params *Params) ([]*asig.AdaptorSignature, error) {
	// construct the map where key is the covenant PK and value is this
	// covenant member's adaptor signature encrypted by the given finality provider's PK
	covSigsMap := map[string]*asig.AdaptorSignature{}
	for _, covSigs := range covSigsList {
		// find the adaptor signature at the corresponding finality provider's index
		if fpIdx >= len(covSigs.AdaptorSigs) {
			return nil, fmt.Errorf("finality provider index is out of the scope")
		}
		covSigBytes := covSigs.AdaptorSigs[fpIdx]
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

// MinimumUnbondingTime returns the minimum unbonding time. It is the bigger value from:
// - MinUnbondingTime
// - CheckpointFinalizationTimeout
func MinimumUnbondingTime(
	stakingParams Params,
	checkpointingParams btcctypes.Params) uint64 {
	return math.Max[uint64](
		uint64(stakingParams.MinUnbondingTime),
		checkpointingParams.CheckpointFinalizationTimeout,
	)
}

func (state *BTCDelegationStatus) ToBytes() []byte {
	return sdk.Uint64ToBigEndian(uint64(*state))
}

func NewBTCDelegationStatus(stateBytes []byte) (BTCDelegationStatus, error) {
	if len(stateBytes) != 8 {
		return -1, fmt.Errorf("malformed bytes for BTC delegation status")
	}
	stateInt := int32(sdk.BigEndianToUint64(stateBytes))
	switch stateInt {
	case 0, 1, 2, 3:
		return BTCDelegationStatus(stateInt), nil
	default:
		return -1, fmt.Errorf("invalid BTC delegation status bytes; should be one of {0, 1, 2, 3}")
	}
}

func NewBTCDelegationStatusFromString(statusStr string) (BTCDelegationStatus, error) {
	switch statusStr {
	case "pending":
		return BTCDelegationStatus_PENDING, nil
	case "active":
		return BTCDelegationStatus_ACTIVE, nil
	case "unbonded":
		return BTCDelegationStatus_UNBONDED, nil
	case "any":
		return BTCDelegationStatus_ANY, nil
	default:
		return -1, fmt.Errorf("invalid status string; should be one of {pending, active, unbonding, unbonded, any}")
	}
}
