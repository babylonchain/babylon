package types

import (
	"bytes"
	"fmt"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

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
		return fmt.Errorf("empty Babylon public key")
	}
	if d.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if d.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if d.ValBtcPk == nil {
		return fmt.Errorf("empty Validator BTC public key")
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
	if err := d.StakingTx.ValidateBasic(); err != nil {
		return err
	}

	// verify PoP
	if err := d.Pop.ValidateBasic(); err != nil {
		return err
	}
	if err := d.Pop.Verify(d.BabylonPk, d.BtcPk); err != nil {
		return err
	}

	return nil
}

// IsActivated returns whether a BTC delegation is activated or not
// a BTC delegation is activated when it receives a signature from jury
func (d *BTCDelegation) IsActivated() bool {
	return d.JurySig != nil
}

// VotingPower returns the voting power of the BTC delegation at a given BTC height
// and a given w value
// The BTC delegation d has voting power iff the following holds
// - d has a jury signature
// - d's timelock start height <= the given BTC height
// - the given BTC height <= d's timelock end height - w
func (d *BTCDelegation) VotingPower(btcHeight uint64, w uint64) uint64 {
	if d.IsActivated() && d.StartHeight <= btcHeight && btcHeight+w <= d.EndHeight {
		return d.TotalSat
	} else {
		return 0
	}
}

func (p *ProofOfPossession) ValidateBasic() error {
	if len(p.BabylonSig) == 0 {
		return fmt.Errorf("empty Babylon signature")
	}
	if p.BtcSig == nil {
		return fmt.Errorf("empty BTC signature")
	}
	if _, err := p.BtcSig.ToBTCSig(); err != nil {
		return fmt.Errorf("BtcSig is incorrectly formatted: %w", err)
	}

	return nil
}

func (tx *StakingTx) Equals(tx2 *StakingTx) bool {
	return bytes.Equal(tx.Tx, tx2.Tx) && bytes.Equal(tx.StakingScript, tx2.StakingScript)
}

func (tx *StakingTx) ValidateBasic() error {
	// unmarshal tx bytes to MsgTx
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(tx.Tx)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return err
	}

	// parse staking script
	if _, err := btcstaking.ParseStakingTransactionScript(tx.StakingScript); err != nil {
		return err
	}

	return nil
}

func (tx *StakingTx) ToMsgTx() (*wire.MsgTx, error) {
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(tx.Tx)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return nil, err
	}
	return &msgTx, nil
}

func (tx *StakingTx) GetStakingScriptData() (*btcstaking.StakingScriptData, error) {
	return btcstaking.ParseStakingTransactionScript(tx.StakingScript)
}

func (tx *StakingTx) GetStakingOutputInfo(net *chaincfg.Params) (*btcstaking.StakingOutputInfo, error) {
	var (
		scriptData *btcstaking.StakingScriptData
		outValue   int64
		err        error
	)

	// unmarshal tx bytes to MsgTx
	var msgTx wire.MsgTx
	rbuf := bytes.NewReader(tx.Tx)
	if err := msgTx.Deserialize(rbuf); err != nil {
		return nil, err
	}

	// parse staking script
	scriptData, err = btcstaking.ParseStakingTransactionScript(tx.StakingScript)
	if err != nil {
		return nil, err
	}
	expectedPkScript, err := btcstaking.BuildUnspendableTaprootPkScript(tx.StakingScript, net)
	if err != nil {
		return nil, err
	}

	// find the output that corresponds to the staking script
	for _, txOut := range msgTx.TxOut {
		if bytes.Equal(expectedPkScript, txOut.PkScript) {
			outValue = txOut.Value
		}
	}
	if outValue == 0 {
		// not found
		return nil, fmt.Errorf("the tx contains no StakingTransactionScript")
	}

	return &btcstaking.StakingOutputInfo{
		StakingScriptData: scriptData,
		StakingPkScript:   expectedPkScript,
		StakingAmount:     btcutil.Amount(outValue),
	}, nil
}
