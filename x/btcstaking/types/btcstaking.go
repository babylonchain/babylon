package types

func (v *BTCValidator) ValidateBasic() error {
	// TODO: validation rules

	if err := v.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (d *BTCDelegation) ValidateBasic() error {
	// TODO: validation rules

	if err := d.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (p *ProofOfPossession) ValidateBasic() error {
	// TODO: validation rules

	return nil
}
