package keeper

import (
	"github.com/babylonchain/babylon/x/headeroracle/types"
)

var _ types.QueryServer = Keeper{}
