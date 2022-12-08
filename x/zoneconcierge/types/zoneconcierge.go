package types

func (p *ProofEpochSealed) ValidateBasic() error {
	if p.ValidatorSet == nil {
		return ErrInvalidProofEpochSealed.Wrap("ValidatorSet is nil")
	} else if len(p.ValidatorSet) == 0 {
		return ErrInvalidProofEpochSealed.Wrap("ValidatorSet is empty")
	} else if p.ProofEpochInfo == nil {
		return ErrInvalidProofEpochSealed.Wrap("ProofEpochInfo is nil")
	} else if p.ProofEpochValSet == nil {
		return ErrInvalidProofEpochSealed.Wrap("ProofEpochValSet is nil")
	}
	return nil
}
