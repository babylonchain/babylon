package keeper

import (
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

var _ types.QueryServer = Keeper{}
