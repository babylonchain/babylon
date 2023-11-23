package types

import (
	"fmt"
	"sort"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func NewBTCDelegationStatusFromString(statusStr string) (BTCDelegationStatus, error) {
	switch statusStr {
	case "pending":
		return BTCDelegationStatus_PENDING, nil
	case "active":
		return BTCDelegationStatus_ACTIVE, nil
	case "unbonding":
		return BTCDelegationStatus_UNBONDING, nil
	case "unbonded":
		return BTCDelegationStatus_UNBONDED, nil
	case "any":
		return BTCDelegationStatus_ANY, nil
	default:
		return -1, fmt.Errorf("invalid status string; should be one of {pending, active, unbonding, unbonded, any}")
	}
}

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

func (d *BTCDelegation) ValidateBasic() error {
	if d.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if d.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if d.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if len(d.ValBtcPkList) == 0 {
		return fmt.Errorf("empty list of BTC validator PKs")
	}
	if ExistsDup(d.ValBtcPkList) {
		return fmt.Errorf("list of BTC validator PKs has duplication")
	}
	if d.StakingTx == nil {
		return fmt.Errorf("empty staking tx")
	}
	if d.SlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if d.DelegatorSig == nil {
		return fmt.Errorf("empty delegator signature")
	}

	// ensure staking tx is correctly formatted
	if _, err := ParseBtcTx(d.StakingTx); err != nil {
		return err
	}
	if err := d.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// HasCovenantSig returns whether a BTC delegation has a covenant signature
func (d *BTCDelegation) HasCovenantQuorum(quorum uint32) bool {
	// TODO: accomodate covenant committee for CovenantSig
	return d.CovenantSig != nil
}

// IsSignedByCovMember checks whether the given covenant PK has signed the delegation
func (d *BTCDelegation) IsSignedByCovMember(covPK *bbn.BIP340PubKey) bool {
	// TODO: accomodate covenant committee for CovenantSig
	return d.CovenantSig != nil
}

func (ud *BTCUndelegation) HasCovenantQuorum(quorum uint32) bool {
	// TODO: accomodate covenant committee for CovenantSlashingSig
	return ud.CovenantSlashingSig != nil && len(ud.CovenantUnbondingSigList) >= int(quorum)
}

// IsSignedByCovMember checks whether the given covenant PK has signed the undelegation
func (ud *BTCUndelegation) IsSignedByCovMember(covPK *bbn.BIP340PubKey) bool {
	// TODO: accomodate covenant committee for CovenantSlashingSig
	for _, sigInfo := range ud.CovenantUnbondingSigList {
		if sigInfo.Pk.Equals(covPK) {
			return true
		}
	}
	return false
}

func (ud *BTCUndelegation) HasAllSignatures(covenantQuorum uint32) bool {
	return ud.HasCovenantQuorum(covenantQuorum)
}

// GetStatus returns the status of the BTC Delegation based on a BTC height and a w value
// TODO: Given that we only accept delegations that can be activated immediately,
// we can only have expired delegations. If we accept optimistic submissions,
// we could also have delegations that are in the waiting, so we need an extra status.
// This is covered by expired for now as it is the default value.
// Active: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation has a covenant sig
// Pending: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation does not have a covenant sig
// Expired: Delegation timelock
func (d *BTCDelegation) GetStatus(btcHeight uint64, w uint64, covenantQuorum uint32) BTCDelegationStatus {
	if d.BtcUndelegation != nil {
		if d.BtcUndelegation.HasAllSignatures(covenantQuorum) {
			return BTCDelegationStatus_UNBONDED
		}
		// If we received an undelegation but is still does not have all required signature,
		// delegation receives UNBONING status.
		// Voting power from this delegation is removed from the total voting power and now we
		// are waiting for signatures from validator and covenant for delegation to become expired.
		// For now we do not have any unbonding time on Babylon chain, only time lock on BTC chain
		// we may consider adding unbonding time on Babylon chain later to avoid situation where
		// we can lose to much voting power in to short time.
		return BTCDelegationStatus_UNBONDING
	}

	if d.StartHeight <= btcHeight && btcHeight+w <= d.EndHeight {
		if d.HasCovenantQuorum(covenantQuorum) {
			return BTCDelegationStatus_ACTIVE
		} else {
			return BTCDelegationStatus_PENDING
		}
	}
	return BTCDelegationStatus_UNBONDED
}

// VotingPower returns the voting power of the BTC delegation at a given BTC height
// and a given w value.
// The BTC delegation d has voting power iff it is active.
func (d *BTCDelegation) VotingPower(btcHeight uint64, w uint64, covenantQuorum uint32) uint64 {
	if d.GetStatus(btcHeight, w, covenantQuorum) != BTCDelegationStatus_ACTIVE {
		return 0
	}
	return d.GetTotalSat()
}

func (d *BTCDelegation) GetStakingTxHash() (chainhash.Hash, error) {
	parsed, err := ParseBtcTx(d.StakingTx)

	if err != nil {
		return chainhash.Hash{}, err
	}

	return parsed.TxHash(), nil
}

func (d *BTCDelegation) MustGetStakingTxHash() chainhash.Hash {
	txHash, err := d.GetStakingTxHash()

	if err != nil {
		panic(err)
	}

	return txHash
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
