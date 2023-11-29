package types

import (
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
)

func (ud *BTCUndelegation) HasCovenantQuorumOnSlashing(quorum uint32) bool {
	return len(ud.CovenantUnbondingSigList) >= int(quorum)
}

func (ud *BTCUndelegation) HasCovenantQuorumOnUnbonding(quorum uint32) bool {
	return len(ud.CovenantUnbondingSigList) >= int(quorum)
}

// IsSignedByCovMemberOnUnbonding checks whether the given covenant PK has signed the unbonding tx
func (ud *BTCUndelegation) IsSignedByCovMemberOnUnbonding(covPK *bbn.BIP340PubKey) bool {
	for _, sigInfo := range ud.CovenantUnbondingSigList {
		if sigInfo.Pk.Equals(covPK) {
			return true
		}
	}
	return false
}

// IsSignedByCovMemberOnSlashing checks whether the given covenant PK has signed the slashing tx
func (ud *BTCUndelegation) IsSignedByCovMemberOnSlashing(covPK *bbn.BIP340PubKey) bool {
	for _, sigInfo := range ud.CovenantSlashingSigs {
		if sigInfo.CovPk.Equals(covPK) {
			return true
		}
	}
	return false
}

func (ud *BTCUndelegation) IsSignedByCovMember(covPk *bbn.BIP340PubKey) bool {
	return ud.IsSignedByCovMemberOnUnbonding(covPk) && ud.IsSignedByCovMemberOnSlashing(covPk)
}

func (ud *BTCUndelegation) HasAllSignatures(covenantQuorum uint32) bool {
	return ud.HasCovenantQuorumOnUnbonding(covenantQuorum) && ud.HasCovenantQuorumOnSlashing(covenantQuorum)
}

// AddCovenantSigs adds a Schnorr signature on the unbonding tx, and
// a list of adaptor signatures on the unbonding slashing tx, each encrypted
// by a BTC validator's PK this BTC delegation restakes to, from the given
// covenant
func (ud *BTCUndelegation) AddCovenantSigs(
	covPk *bbn.BIP340PubKey,
	unbondingSig *bbn.BIP340Signature,
	slashingSigs []asig.AdaptorSignature,
	quorum uint32,
) error {
	// we can ignore the covenant slashing sig if quorum is already reached
	if ud.HasAllSignatures(quorum) {
		return nil
	}

	if ud.IsSignedByCovMember(covPk) {
		return ErrDuplicatedCovenantSig
	}

	covUnbondingSigInfo := &SignatureInfo{Pk: covPk, Sig: unbondingSig}
	ud.CovenantUnbondingSigList = append(ud.CovenantUnbondingSigList, covUnbondingSigInfo)

	adaptorSigs := make([][]byte, 0, len(slashingSigs))
	for _, s := range slashingSigs {
		adaptorSigs = append(adaptorSigs, s.MustMarshal())
	}
	slashingSigsInfo := &CovenantAdaptorSignatures{CovPk: covPk, AdaptorSigs: adaptorSigs}
	ud.CovenantSlashingSigs = append(ud.CovenantSlashingSigs, slashingSigsInfo)

	return nil
}
