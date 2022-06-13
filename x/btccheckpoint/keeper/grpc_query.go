package keeper

import (
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

var _ types.QueryServer = Keeper{}
