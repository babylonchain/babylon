package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

var _ types.QueryServer = Keeper{}
