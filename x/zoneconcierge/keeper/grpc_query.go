package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
)

var _ types.QueryServer = Keeper{}
