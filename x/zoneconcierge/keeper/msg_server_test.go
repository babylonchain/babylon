package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/babylonchain/babylon/x/zoneconcierge/types"
    "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
    keepertest "github.com/babylonchain/babylon/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.ZoneconciergeKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
