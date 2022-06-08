package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
)

var _ types.QueryServer = Keeper{}
