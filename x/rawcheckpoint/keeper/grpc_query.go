package keeper

import (
	"github.com/babylonchain/babylon/x/rawcheckpoint/types"
)

var _ types.QueryServer = Keeper{}
