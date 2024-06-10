package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
)

func headerInfoFromStoredBytes(cdc codec.BinaryCodec, bz []byte) *types.BTCHeaderInfo {
	headerInfo := new(types.BTCHeaderInfo)
	cdc.MustUnmarshal(bz, headerInfo)
	return headerInfo
}

// Logger returns the logger with the key value of the current module.
func Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// emitTypedEventWithLog emits an event and logs if it errors.
func emitTypedEventWithLog(ctx context.Context, evt proto.Message) {
	if err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(evt); err != nil {
		Logger(sdk.UnwrapSDKContext(ctx)).Error(
			"faied to emit event",
			"type", evt.String(),
			"reason", err.Error(),
		)
	}
}
