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

// GetFpIdx returns the index of the finality provider in the list of finality providers
// that the BTC delegation is restaked to
func (d *BTCDelegation) GetFpIdx(fpBTCPK *bbn.BIP340PubKey) int {
	for i := 0; i < len(d.FpBtcPkList); i++ {
		if d.FpBtcPkList[i].Equals(fpBTCPK) {
			return i
		}
	}
	return -1
}

func (d *BTCDelegation) GetCovSlashingAdaptorSig(
	covBTCPK *bbn.BIP340PubKey,
	valIdx int,
	quorum uint32,
) (*asig.AdaptorSignature, error) {
	if !d.HasCovenantQuorums(quorum) {
		return nil, ErrInvalidDelegationState.Wrap("BTC delegation does not have a covenant quorum yet")
	}
	for _, covASigs := range d.CovenantSigs {
		if covASigs.CovPk.Equals(covBTCPK) {
			if valIdx >= len(covASigs.AdaptorSigs) {
				return nil, ErrFpNotFound.Wrap("validator index is out of scope")
			}
			sigBytes := covASigs.AdaptorSigs[valIdx]
			return asig.NewAdaptorSignatureFromBytes(sigBytes)
		}
	}

	return nil, ErrInvalidCovenantPK.Wrap("covenant PK is not found")
}

// IsUnbondedEarly returns whether the delegator has signed unbonding signature.
// Signing unbonding signature means the delegator wants to unbond early, and
// Babylon will consider this BTC delegation unbonded directly
func (d *BTCDelegation) IsUnbondedEarly() bool {
	return d.BtcUndelegation.DelegatorUnbondingSig != nil
}

// GetStatus returns the status of the BTC Delegation based on BTC height, w value, and covenant quorum
// Pending: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation does not have covenant signatures
// Active: the BTC height is in the range of d's [startHeight, endHeight-w] and the delegation has quorum number of signatures over slashing tx, unbonding tx, and slashing unbonding tx from covenant committee
// Unbonded: the BTC height is larger than `endHeight-w` or the BTC delegation has received a signature on unbonding tx from the delegator
func (d *BTCDelegation) GetStatus(btcHeight uint64, w uint64, covenantQuorum uint32) BTCDelegationStatus {
	if d.IsUnbondedEarly() {
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
	if len(d.FpBtcPkList) == 0 {
		return fmt.Errorf("empty list of finality provider PKs")
	}
	if ExistsDup(d.FpBtcPkList) {
		return fmt.Errorf("list of finality provider PKs has duplication")
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
// each finality provider's PK this BTC delegation restakes to
// It is up to the caller to ensure that given adaptor signatures are valid or
// that they were not added before
func (d *BTCDelegation) AddCovenantSigs(
	covPk *bbn.BIP340PubKey,
	stakingSlashingSigs []asig.AdaptorSignature,
	unbondingSig *bbn.BIP340Signature,
	unbondingSlashingSigs []asig.AdaptorSignature,
) {
	adaptorSigs := make([][]byte, 0, len(stakingSlashingSigs))
	for _, s := range stakingSlashingSigs {
		adaptorSigs = append(adaptorSigs, s.MustMarshal())
	}
	covSigs := &CovenantAdaptorSignatures{CovPk: covPk, AdaptorSigs: adaptorSigs}

	d.CovenantSigs = append(d.CovenantSigs, covSigs)
	// add unbonding sig and unbonding slashing adaptor sig
	d.BtcUndelegation.addCovenantSigs(covPk, unbondingSig, unbondingSlashingSigs)
}

// GetStakingInfo returns the staking info of the BTC delegation
// the staking info can be used for constructing witness of slashing tx
// with access to a finality provider's SK
func (d *BTCDelegation) GetStakingInfo(bsParams *Params, btcNet *chaincfg.Params) (*btcstaking.StakingInfo, error) {
	fpBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(d.FpBtcPkList)
	if err != nil {
		return nil, fmt.Errorf("failed to convert finality provider pks to BTC pks %v", err)
	}
	covenantBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	if err != nil {
		return nil, fmt.Errorf("failed to convert covenant pks to BTC pks %v", err)
	}
	stakingInfo, err := btcstaking.BuildStakingInfo(
		d.BtcPk.MustToBTCPK(),
		fpBtcPkList,
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
// tx with access to a finality provider's SK
func (d *BTCDelegation) GetUnbondingInfo(bsParams *Params, btcNet *chaincfg.Params) (*btcstaking.UnbondingInfo, error) {
	fpBtcPkList, err := bbn.NewBTCPKsFromBIP340PKs(d.FpBtcPkList)
	if err != nil {
		return nil, fmt.Errorf("failed to convert finality provider pks to BTC pks: %v", err)
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
		fpBtcPkList,
		covenantBtcPkList,
		bsParams.CovenantQuorum,
		uint16(d.GetUnbondingTime()),
		btcutil.Amount(unbondingTx.TxOut[0].Value),
		btcNet,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create BTC staking info: %v", err)
	}

	return unbondingInfo, nil
}

// TODO: verify to remove, not used in babylon, only for tests
// findFPIdx returns the index of the given finality provider
// among all restaked finality providers
func (d *BTCDelegation) findFPIdx(fpBTCPK *bbn.BIP340PubKey) (int, error) {
	for i, pk := range d.FpBtcPkList {
		if pk.Equals(fpBTCPK) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("the given finality provider's PK is not found in the BTC delegation")
}

// BuildSlashingTxWithWitness uses the given finality provider's SK to complete
// the signatures on the slashing tx, such that the slashing tx obtains full
// witness and can be submitted to Bitcoin.
// This happens after the finality provider is slashed and its SK is extracted.
// TODO: verify not used
func (d *BTCDelegation) BuildSlashingTxWithWitness(bsParams *Params, btcNet *chaincfg.Params, fpSK *btcec.PrivateKey) (*wire.MsgTx, error) {
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

	// get the list of covenant signatures encrypted by the given finality provider's PK
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(fpSK.PubKey())
	fpIdx, err := d.findFPIdx(fpBTCPK)
	if err != nil {
		return nil, err
	}
	covAdaptorSigs, err := GetOrderedCovenantSignatures(fpIdx, d.CovenantSigs, bsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get ordered covenant adaptor signatures: %w", err)
	}

	// assemble witness for slashing tx
	slashingMsgTxWithWitness, err := d.SlashingTx.BuildSlashingTxWithWitness(
		fpSK,
		d.FpBtcPkList,
		stakingMsgTx,
		d.StakingOutputIdx,
		d.DelegatorSig,
		covAdaptorSigs,
		slashingSpendInfo,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to build witness for BTC delegation of %s under finality provider %s: %v",
			d.BtcPk.MarshalHex(),
			bbn.NewBIP340PubKeyFromBTCPK(fpSK.PubKey()).MarshalHex(),
			err,
		)
	}

	return slashingMsgTxWithWitness, nil
}

// TODO: verify to remove, func not used by babylon, used in side car processes.
func (d *BTCDelegation) BuildUnbondingSlashingTxWithWitness(bsParams *Params, btcNet *chaincfg.Params, fpSK *btcec.PrivateKey) (*wire.MsgTx, error) {
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

	// get the list of covenant signatures encrypted by the given finality provider's PK
	fpPK := fpSK.PubKey()
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(fpPK)
	fpIdx, err := d.findFPIdx(fpBTCPK)
	if err != nil {
		return nil, err
	}
	covAdaptorSigs, err := GetOrderedCovenantSignatures(fpIdx, d.BtcUndelegation.CovenantSlashingSigs, bsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get ordered covenant adaptor signatures: %w", err)
	}

	// assemble witness for unbonding slashing tx
	slashingMsgTxWithWitness, err := d.BtcUndelegation.SlashingTx.BuildSlashingTxWithWitness(
		fpSK,
		d.FpBtcPkList,
		unbondingMsgTx,
		0,
		d.BtcUndelegation.DelegatorSlashingSig,
		covAdaptorSigs,
		slashingSpendInfo,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to build witness for unbonding BTC delegation %s under finality provider %s: %v",
			d.BtcPk.MarshalHex(),
			bbn.NewBIP340PubKeyFromBTCPK(fpSK.PubKey()).MarshalHex(),
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
