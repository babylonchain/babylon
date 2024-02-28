package types

import (
	"encoding/hex"
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
		DelegatorSlashSigHex: btcDel.DelegatorSig.ToHexStr(),
		CovenantSigs:         btcDel.CovenantSigs,
		Active:               status == BTCDelegationStatus_ACTIVE,
		StatusDesc:           status.String(),
		UnbondingTime:        btcDel.UnbondingTime,
		UndelegationResponse: nil,
	}

	if btcDel.SlashingTx != nil {
		resp.SlashingTxHex = hex.EncodeToString(*btcDel.SlashingTx)
	}

	if btcDel.BtcUndelegation != nil {
		resp.UndelegationResponse = btcDel.BtcUndelegation.ToResponse()
	}

	return resp
}

// ToResponse parses an BTCUndelegation into BTCUndelegationResponse.
func (ud *BTCUndelegation) ToResponse() (resp *BTCUndelegationResponse) {
	resp = &BTCUndelegationResponse{
		UnbondingTxHex:           hex.EncodeToString(ud.UnbondingTx),
		CovenantUnbondingSigList: ud.CovenantUnbondingSigList,
		CovenantSlashingSigs:     ud.CovenantSlashingSigs,
	}

	if ud.DelegatorUnbondingSig != nil {
		resp.DelegatorUnbondingSigHex = ud.DelegatorUnbondingSig.ToHexStr()
	}
	if ud.SlashingTx != nil {
		resp.SlashingTxHex = ud.SlashingTx.ToHexStr()
	}
	if ud.DelegatorSlashingSig != nil {
		resp.DelegatorSlashingSigHex = ud.DelegatorSlashingSig.ToHexStr()
	}

	return resp
}
