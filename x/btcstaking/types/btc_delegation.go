package types

import (
	"bytes"
	"fmt"
	math "math"

	"github.com/babylonchain/babylon/btcstaking"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

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

func (d *BTCDelegation) GetStakingTime() uint16 {
	diff := d.EndHeight - d.StartHeight

	if diff > math.MaxUint16 {
		// In valid delegation, EndHeight is always greater than StartHeight and it is always uint16 value
		panic("invalid delegation in database")
	}

	return uint16(diff)
}

// GetStatus returns the status of the BTC Delegation based on a BTC height and a w value
// TODO: Given that we only accept delegations that can be activated immediately,
// we can only have expired delegations. If we accept optimistic submissions,
// we could also have delegations that are in the waiting, so we need an extra status.
// This is covered by expired for now as it is the default value.
// Active: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation has a covenant sig
// Pending: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation does not have a covenant sig
// TODO: fix comment above
// Expired: Delegation timelock
func (d *BTCDelegation) GetStatus(btcHeight uint64, w uint64, covenantQuorum uint32) BTCDelegationStatus {
	if d.BtcUndelegation.DelegatorUnbondingSig != nil {
		// this means the delegator has signed unbonding signature, and Babylon will consider
		// this BTC delegation unbonded directly
		return BTCDelegationStatus_UNBONDED
	}

	if btcHeight < d.StartHeight || btcHeight+w > d.EndHeight {
		// staking tx's timelock has not begun, or is less than w BTC
		// blocks left, or is expired
		return BTCDelegationStatus_UNBONDED
	}

	// at this point, BTC delegation has an active timelock, and Babylon is not
	// aware of unbonding tx with delegator's signature
	if d.HasCovenantQuorums(covenantQuorum) {
		// this BTC delegation receives covenant quorums on
		// {slashing/unbonding/unbondingslashing} txs, thus is active
		return BTCDelegationStatus_ACTIVE
	}

	// no covenant quorum yet, pending
	return BTCDelegationStatus_PENDING
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
	parsed, err := bbn.NewBTCTxFromBytes(d.StakingTx)

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
	if _, err := bbn.NewBTCTxFromBytes(d.StakingTx); err != nil {
		return err
	}
	if err := d.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// HasCovenantQuorum returns whether a BTC delegation has a quorum number of signatures
// from covenant members, including
// - adaptor signatures on slashing tx
// - Schnorr signatures on unbonding tx
// - adaptor signatrues on unbonding slashing tx
func (d *BTCDelegation) HasCovenantQuorums(quorum uint32) bool {
	return uint32(len(d.CovenantSigs)) >= quorum && d.BtcUndelegation.HasCovenantQuorums(quorum)
}

// IsSignedByCovMember checks whether the given covenant PK has signed the delegation
func (d *BTCDelegation) IsSignedByCovMember(covPk *bbn.BIP340PubKey) bool {
	for _, sigInfo := range d.CovenantSigs {
		if covPk.Equals(sigInfo.CovPk) {
			return true
		}
	}

	return false
}

// AddCovenantSigs adds signatures on the slashing tx from the given
// covenant, where each signature is an adaptor signature encrypted by
// each BTC validator's PK this BTC delegation restakes to
func (d *BTCDelegation) AddCovenantSigs(covPk *bbn.BIP340PubKey, sigs []asig.AdaptorSignature, quorum uint32) error {
	// we can ignore the covenant sig if quorum is already reached
	if d.HasCovenantQuorums(quorum) {
		return nil
	}
	// ensure that this covenant member has not signed the delegation yet
	if d.IsSignedByCovMember(covPk) {
		return ErrDuplicatedCovenantSig
	}

	adaptorSigs := make([][]byte, 0, len(sigs))
	for _, s := range sigs {
		adaptorSigs = append(adaptorSigs, s.MustMarshal())
	}
	covSigs := &CovenantAdaptorSignatures{CovPk: covPk, AdaptorSigs: adaptorSigs}

	d.CovenantSigs = append(d.CovenantSigs, covSigs)

	return nil
}

// GetStakingInfo returns the staking info of the BTC delegation
// the staking info can be used for constructing witness of slashing tx
// with access to a BTC validator's SK
func (d *BTCDelegation) GetStakingInfo(bsParams *Params, btcNet *chaincfg.Params) (*btcstaking.StakingInfo, error) {
	valBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(d.ValBtcPkList)
	if err != nil {
		return nil, fmt.Errorf("failed to convert validator pks to BTC pks %v", err)
	}
	covenantBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	if err != nil {
		return nil, fmt.Errorf("failed to convert covenant pks to BTC pks %v", err)
	}
	stakingInfo, err := btcstaking.BuildStakingInfo(
		d.BtcPk.MustToBTCPK(),
		valBtcPkList,
		covenantBtcPkList,
		bsParams.CovenantQuorum,
		d.GetStakingTime(),
		btcutil.Amount(d.TotalSat),
		btcNet,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create BTC staking info: %v", err)
	}
	return stakingInfo, nil
}

func (d *BTCDelegation) SignUnbondingTx(bsParams *Params, btcNet *chaincfg.Params, sk *btcec.PrivateKey) (*schnorr.Signature, error) {
	stakingTx, err := bbn.NewBTCTxFromBytes(d.StakingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse staking transaction: %v", err)
	}
	unbondingTx, err := bbn.NewBTCTxFromBytes(d.BtcUndelegation.UnbondingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unbonding transaction: %v", err)
	}
	stakingInfo, err := d.GetStakingInfo(bsParams, btcNet)
	if err != nil {
		return nil, err
	}
	unbondingPath, err := stakingInfo.UnbondingPathSpendInfo()
	if err != nil {
		return nil, err
	}

	sig, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
		unbondingTx,
		stakingTx,
		d.StakingOutputIdx,
		unbondingPath.GetPkScriptPath(),
		sk,
	)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// GetUnbondingInfo returns the unbonding info of the BTC delegation
// the unbonding info can be used for constructing witness of unbonding slashing
// tx with access to a BTC validator's SK
func (d *BTCDelegation) GetUnbondingInfo(bsParams *Params, btcNet *chaincfg.Params) (*btcstaking.UnbondingInfo, error) {
	valBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(d.ValBtcPkList)
	if err != nil {
		return nil, fmt.Errorf("failed to convert validator pks to BTC pks: %v", err)
	}

	covenantBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	if err != nil {
		return nil, fmt.Errorf("failed to convert covenant pks to BTC pks: %v", err)
	}
	unbondingTx, err := bbn.NewBTCTxFromBytes(d.BtcUndelegation.UnbondingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unbonding transaction: %v", err)
	}

	unbondingInfo, err := btcstaking.BuildUnbondingInfo(
		d.BtcPk.MustToBTCPK(),
		valBtcPkList,
		covenantBtcPkList,
		bsParams.CovenantQuorum,
		uint16(d.BtcUndelegation.GetUnbondingTime()),
		btcutil.Amount(unbondingTx.TxOut[0].Value),
		btcNet,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create BTC staking info: %v", err)
	}

	return unbondingInfo, nil
}

func (d *BTCDelegation) BuildSlashingTxWithWitness(bsParams *Params, btcNet *chaincfg.Params, valSK *btcec.PrivateKey) (*wire.MsgTx, error) {
	stakingMsgTx, err := bbn.NewBTCTxFromBytes(d.StakingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert a Babylon staking tx to wire.MsgTx: %w", err)
	}

	// get staking info
	stakingInfo, err := d.GetStakingInfo(bsParams, btcNet)
	if err != nil {
		return nil, fmt.Errorf("could not create BTC staking info: %v", err)
	}
	slashingSpendInfo, err := stakingInfo.SlashingPathSpendInfo()
	if err != nil {
		return nil, fmt.Errorf("could not get slashing spend info: %v", err)
	}

	// TODO: work with restaking
	covAdaptorSigs, err := GetOrderedCovenantSignatures(0, d.CovenantSigs, bsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get ordered covenant adaptor signatures: %w", err)
	}

	// assemble witness for slashing tx
	slashingMsgTxWithWitness, err := d.SlashingTx.BuildSlashingTxWithWitness(
		valSK,
		stakingMsgTx,
		d.StakingOutputIdx,
		d.DelegatorSig,
		covAdaptorSigs,
		slashingSpendInfo,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to build witness for BTC delegation of %s under BTC validator %s: %v",
			d.BtcPk.MarshalHex(),
			bbn.NewBIP340PubKeyFromBTCPK(valSK.PubKey()).MarshalHex(),
			err,
		)
	}

	return slashingMsgTxWithWitness, nil
}

func (d *BTCDelegation) BuildUnbondingSlashingTxWithWitness(bsParams *Params, btcNet *chaincfg.Params, valSK *btcec.PrivateKey) (*wire.MsgTx, error) {
	unbondingMsgTx, err := bbn.NewBTCTxFromBytes(d.BtcUndelegation.UnbondingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert a Babylon unbonding tx to wire.MsgTx: %w", err)
	}

	// get unbonding info
	unbondingInfo, err := d.GetUnbondingInfo(bsParams, btcNet)
	if err != nil {
		return nil, fmt.Errorf("could not create BTC unbonding info: %v", err)
	}
	slashingSpendInfo, err := unbondingInfo.SlashingPathSpendInfo()
	if err != nil {
		return nil, fmt.Errorf("could not get unbonding slashing spend info: %v", err)
	}

	// TODO: work with restaking
	covAdaptorSigs, err := GetOrderedCovenantSignatures(0, d.BtcUndelegation.CovenantSlashingSigs, bsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get ordered covenant adaptor signatures: %w", err)
	}

	// assemble witness for unbonding slashing tx
	slashingMsgTxWithWitness, err := d.BtcUndelegation.SlashingTx.BuildSlashingTxWithWitness(
		valSK,
		unbondingMsgTx,
		0,
		d.BtcUndelegation.DelegatorSlashingSig,
		covAdaptorSigs,
		slashingSpendInfo,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to build witness for unbonding BTC delegation %s under BTC validator %s: %v",
			d.BtcPk.MarshalHex(),
			bbn.NewBIP340PubKeyFromBTCPK(valSK.PubKey()).MarshalHex(),
			err,
		)
	}

	return slashingMsgTxWithWitness, nil
}

func NewBTCDelegatorDelegationIndex() *BTCDelegatorDelegationIndex {
	return &BTCDelegatorDelegationIndex{
		StakingTxHashList: [][]byte{},
	}
}

func (i *BTCDelegatorDelegationIndex) Has(stakingTxHash chainhash.Hash) bool {
	for _, hash := range i.StakingTxHashList {
		if bytes.Equal(stakingTxHash[:], hash) {
			return true
		}
	}
	return false
}

func (i *BTCDelegatorDelegationIndex) Add(stakingTxHash chainhash.Hash) error {
	// ensure staking tx hash is not duplicated
	for _, hash := range i.StakingTxHashList {
		if bytes.Equal(stakingTxHash[:], hash) {
			return fmt.Errorf("the given stakingTxHash %s is duplicated", stakingTxHash.String())
		}
	}
	// add
	i.StakingTxHashList = append(i.StakingTxHashList, stakingTxHash[:])

	return nil
}

// VotingPower calculates the total voting power of all BTC delegations
func (dels *BTCDelegatorDelegations) VotingPower(btcHeight uint64, w uint64, covenantQuorum uint32) uint64 {
	power := uint64(0)
	for _, del := range dels.Dels {
		power += del.VotingPower(btcHeight, w, covenantQuorum)
	}
	return power
}
