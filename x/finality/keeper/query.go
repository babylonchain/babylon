package keeper

import (
	"github.com/babylonchain/babylon/x/finality/types"
)

var _ types.QueryServer = Keeper{}
