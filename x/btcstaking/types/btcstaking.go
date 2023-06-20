package types

import "fmt"

func (v *BTCValidator) ValidateBasic() error {
	// ensure fields are non-empty and well-formatted
	if v.BabylonPk == nil {
		return fmt.Errorf("empty BabylonPk")
	}
	if v.BtcPk == nil {
		return fmt.Errorf("empty BtcPk")
	}
	if _, err := v.BtcPk.ToBTCPK(); err != nil {
		return fmt.Errorf("BtcPk is not correctly formatted: %w", err)
	}
	if v.Pop == nil {
		return fmt.Errorf("empty Pop")
	}

	// verify PoP
	if err := v.Pop.ValidateBasic(); err != nil {
		return err
	}
	if err := v.Pop.Verify(v.BabylonPk, v.BtcPk); err != nil {
		return err
	}

	return nil
}

func (d *BTCDelegation) ValidateBasic() error {
	if d.BabylonPk == nil {
		return fmt.Errorf("empty BabylonPk")
	}
	if d.BtcPk == nil {
		return fmt.Errorf("empty BabylonPk")
	}
	if d.Pop == nil {
		return fmt.Errorf("empty Pop")
	}
	if d.ValBtcPk == nil {
		return fmt.Errorf("empty ValBtcPk")
	}
	if d.StakingTx == nil {
		return fmt.Errorf("empty StakingTx")
	}
	if d.StakingTxSig == nil {
		return fmt.Errorf("empty StakingTxSig")
	}
	if d.StakingTx == nil {
		return fmt.Errorf("empty StakingTxInfo")
	}
	if d.SlashingTx == nil {
		return fmt.Errorf("empty SlashingTx")
	}
	if d.SlashingTxSig == nil {
		return fmt.Errorf("empty SlashingTxSig")
	}

	// TODO: validation rules

	// verify PoP
	if err := d.Pop.ValidateBasic(); err != nil {
		return err
	}
	if err := d.Pop.Verify(d.BabylonPk, d.BtcPk); err != nil {
		return err
	}

	return nil
}

func (p *ProofOfPossession) ValidateBasic() error {
	if len(p.BabylonSig) == 0 {
		return fmt.Errorf("empty BabylonSig")
	}
	if p.BtcSig == nil {
		return fmt.Errorf("empty BtcSig")
	}
	if _, err := p.BtcSig.ToBTCSig(); err != nil {
		return fmt.Errorf("BtcSig is incorrectly formatted: %w", err)
	}

	return nil
}
