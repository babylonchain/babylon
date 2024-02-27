package types

import "encoding/hex"

// NewBTCDelegationResponse returns a new delegation response structure.
func NewBTCDelegationResponse(btcDel *BTCDelegation, status BTCDelegationStatus) *BTCDelegationResponse {
	return &BTCDelegationResponse{
		BtcPk:         btcDel.BtcPk,
		FpBtcPkList:   btcDel.FpBtcPkList,
		StartHeight:   btcDel.StartHeight,
		EndHeight:     btcDel.EndHeight,
		TotalSat:      btcDel.TotalSat,
		StakingTxHex:  hex.EncodeToString(btcDel.StakingTx),
		CovenantSigs:  btcDel.CovenantSigs,
		Active:        status == BTCDelegationStatus_ACTIVE,
		StatusDesc:    status.String(),
		UnbondingTime: btcDel.UnbondingTime,
		UndelegationInfo: &BTCUndelegationInfo{
			UnbondingTx:              btcDel.BtcUndelegation.UnbondingTx,
			CovenantUnbondingSigList: btcDel.BtcUndelegation.CovenantUnbondingSigList,
			CovenantSlashingSigs:     btcDel.BtcUndelegation.CovenantSlashingSigs,
		},
	}
}
