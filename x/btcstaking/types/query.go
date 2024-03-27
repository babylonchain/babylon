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
		StakingOutputIdx:     btcDel.StakingOutputIdx,
		Active:               status == BTCDelegationStatus_ACTIVE,
		StatusDesc:           status.String(),
		UnbondingTime:        btcDel.UnbondingTime,
		UndelegationResponse: nil,
		ParamsVersion:        btcDel.ParamsVersion,
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

// NewFinalityProviderResponse creates a new finality provider response based on the finaliny provider and his voting power.
func NewFinalityProviderResponse(f *FinalityProvider, bbnBlockHeight, votingPower uint64) *FinalityProviderResponse {
	return &FinalityProviderResponse{
		Description:          f.Description,
		Commission:           f.Commission,
		BabylonPk:            f.BabylonPk,
		BtcPk:                f.BtcPk,
		Pop:                  f.Pop,
		SlashedBabylonHeight: f.SlashedBabylonHeight,
		SlashedBtcHeight:     f.SlashedBtcHeight,
		Height:               bbnBlockHeight,
		VotingPower:          votingPower,
	}
}
