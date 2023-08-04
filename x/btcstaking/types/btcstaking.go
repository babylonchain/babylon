package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
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

// HasJurySig returns whether a BTC delegation has a jury signature
func (d *BTCDelegation) HasJurySig() bool {
	return d.JurySig != nil
}

// GetStatus returns the status of the BTC Delegation based on a BTC height and a w value
// TODO: Given that we only accept delegations that can be activated immediately,
// we can only have expired delegations. If we accept optimistic submissions,
// we could also have delegations that are in the waiting, so we need an extra status.
// This is covered by expired for now as it is the default value.
// Active: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation has a jury sig
// Pending: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation does not have a jury sig
// Expired: Delegation timelock
func (d *BTCDelegation) GetStatus(btcHeight uint64, w uint64) BTCDelegationStatus {
	if d.StartHeight <= btcHeight && btcHeight+w <= d.EndHeight {
		if d.HasJurySig() {
			return BTCDelegationStatus_ACTIVE
		} else {
			return BTCDelegationStatus_PENDING
		}
	}
	return BTCDelegationStatus_EXPIRED
}

// VotingPower returns the voting power of the BTC delegation at a given BTC height
// and a given w value.
// The BTC delegation d has voting power iff it is active.
func (d *BTCDelegation) VotingPower(btcHeight uint64, w uint64) uint64 {
	if d.GetStatus(btcHeight, w) != BTCDelegationStatus_ACTIVE {
		return 0
	}
	return d.GetTotalSat()
}

// GetStakingTxHash returns the staking tx hash of the BTC delegation
// it can be used for uniquely identifying a BTC delegation
func (d *BTCDelegation) GetStakingTxHash() (string, error) {
	return d.StakingTx.GetTxHash()
}

func (d *BTCDelegation) MustGetStakingTxHash() string {
	return d.StakingTx.MustGetTxHash()
}

func NewBTCDelegations() *BTCDelegations {
	return &BTCDelegations{
		Dels: []*BTCDelegation{},
	}
}

// Add appends a given BTC delegation to the BTC delegations
// It requires the given BTC delegation is not in the list yet
// TODO: this is an O(n) operation. Consider optimisation later
func (dels *BTCDelegations) Add(del *BTCDelegation) error {
	stakingTxHash, err := del.GetStakingTxHash()
	if err != nil {
		return fmt.Errorf("failed to add BTC delegation to BTC delegations: %w", err)
	}
	// ensure the given del is not duplicated
	if dels.Has(stakingTxHash) {
		return fmt.Errorf("the given BTC delegation %s is duplicated", stakingTxHash)
	}
	// append
	dels.Dels = append(dels.Dels, del)
	return nil
}

// AddJurySig adds a jury signature to an existing BTC delegation in the BTC delegations
// TODO: this is an O(n) operation. Consider optimisation later
func (dels *BTCDelegations) AddJurySig(stakingTxHash string, sig *bbn.BIP340Signature) error {
	del, err := dels.Get(stakingTxHash)
	if err != nil {
		return fmt.Errorf("cannot find the BTC delegation with staking tx hash %s: %w", stakingTxHash, err)
	}
	if del.JurySig != nil {
		return fmt.Errorf("the BTC delegation with staking tx hash %s already has a jury signature", stakingTxHash)
	}
	del.JurySig = sig
	return nil
}

// TODO: this is an O(n) operation. Consider optimisation later
func (dels *BTCDelegations) Has(stakingTxHash string) bool {
	for _, d := range dels.Dels {
		dStakingTxHash := d.MustGetStakingTxHash()
		if dStakingTxHash == stakingTxHash {
			return true
		}
	}
	return false
}

// TODO: this is an O(n) operation. Consider optimisation later
func (dels *BTCDelegations) Get(stakingTxHash string) (*BTCDelegation, error) {
	for _, d := range dels.Dels {
		dStakingTxHash := d.MustGetStakingTxHash()
		if dStakingTxHash == stakingTxHash {
			return d, nil
		}
	}
	return nil, fmt.Errorf("cannot find the BTC delegation with staking tx hash %s", stakingTxHash)
}

// VotingPower calculates the total voting power of all BTC delegations
func (dels *BTCDelegations) VotingPower(btcHeight uint64, w uint64) uint64 {
	power := uint64(0)
	for _, del := range dels.Dels {
		power += del.VotingPower(btcHeight, w)
	}
	return power
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

func NewStakingTxFromHex(txHex string) (*StakingTx, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}
	var tx StakingTx
	if err := tx.Unmarshal(txBytes); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (tx *StakingTx) ToHexStr() (string, error) {
	txBytes, err := tx.Marshal()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(txBytes), nil
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

func (tx *StakingTx) GetTxHash() (string, error) {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		return "", err
	}
	return msgTx.TxHash().String(), nil
}

func (tx *StakingTx) MustGetTxHash() string {
	msgTx, err := tx.ToMsgTx()
	if err != nil {
		panic(err)
	}
	return msgTx.TxHash().String()
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
