package types

import (
	bbntypes "github.com/babylonchain/babylon/types"
)

// NewFinalityProviderSigningInfo creates a new FinalityProviderSigningInfo instance
func NewFinalityProviderSigningInfo(
	fpPk *bbntypes.BIP340PubKey, startHeight, missedBlocksCounter int64,
) FinalityProviderSigningInfo {
	return FinalityProviderSigningInfo{
		FpBtcPk:             fpPk,
		StartHeight:         startHeight,
		MissedBlocksCounter: missedBlocksCounter,
	}
}
