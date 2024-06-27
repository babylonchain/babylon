package types

import (
	"time"

	bbntypes "github.com/babylonchain/babylon/types"
)

// NewFinalityProviderSigningInfo creates a new FinalityProviderSigningInfo instance
func NewFinalityProviderSigningInfo(
	fpPk *bbntypes.BIP340PubKey, startHeight int64,
	jailedUntil time.Time, missedBlocksCounter int64,
) FinalityProviderSigningInfo {
	return FinalityProviderSigningInfo{
		FpBtcPk:             fpPk,
		StartHeight:         startHeight,
		JailedUntil:         jailedUntil,
		MissedBlocksCounter: missedBlocksCounter,
	}
}
