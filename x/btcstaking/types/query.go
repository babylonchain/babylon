package types

import (
	"encoding/hex"

	types "github.com/babylonchain/babylon/types"
)

// NewBTCDelegationResponse returns a new delegation response structure.
func NewBTCDelegationResponse(btcDel *BTCDelegation, status BTCDelegationStatus) (resp *BTCDelegationResponse) {
	resp = &BTCDelegationResponse{
		BtcPk:                btcDel.BtcPk,
		FpBtcPkList:          btcDel.FpBtcPkList,
		StartHeight:          btcDel.StartHeight,
		EndHeight:            btcDel.EndHeight,
		TotalSat:             btcDel.TotalSat,
		StakingTxHex:         hex.EncodeToString(btcDel.StakingTx),
		SlashingTxHex:        hex.EncodeToString(*btcDel.SlashingTx),
		DelegatorSlashSigHex: btcDel.DelegatorSig.ToHexStr(),
		CovenantSigs:         btcDel.CovenantSigs,
		Active:               status == BTCDelegationStatus_ACTIVE,
		StatusDesc:           status.String(),
		UnbondingTime:        btcDel.UnbondingTime,
		UndelegationResponse: nil,
	}

	if btcDel.BtcUndelegation == nil {
		return resp
	}
	resp.UndelegationResponse = btcDel.BtcUndelegation.ToResponse()

	return resp
}

// ToResponse parses an BTCUndelegation into BTCUndelegationResponse.
func (ud *BTCUndelegation) ToResponse() (resp *BTCUndelegationResponse) {
	resp = &BTCUndelegationResponse{
		UnbondingTxHex:           hex.EncodeToString(ud.UnbondingTx),
		DelegatorUnbondingSigHex: ud.DelegatorUnbondingSig.ToHexStr(),
		CovenantUnbondingSigList: ud.CovenantUnbondingSigList,

		CovenantSlashingSigs: ud.CovenantSlashingSigs,
	}

	if ud.SlashingTx == nil {
		return resp
	}
	slashSig := types.BIP340Signature(*ud.SlashingTx)
	resp.SlashingTxHex = slashSig.ToHexStr()
	resp.DelegatorSlashingSigHex = ud.DelegatorSlashingSig.ToHexStr()

	return resp
}
