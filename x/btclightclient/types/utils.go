package types

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
)

func BlocksPerRetarget(params *chaincfg.Params) int32 {
	targetTimespan := int64(params.TargetTimespan / time.Second)
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)
	return int32(targetTimespan / targetTimePerBlock)
}

func IsRetargetBlock(info *BTCHeaderInfo, params *chaincfg.Params) bool {
	hI32 := int32(info.Height)
	return hI32%BlocksPerRetarget(params) == 0
}
