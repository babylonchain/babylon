package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateValidatorBLS(goCtx context.Context, msg *types.MsgCreateValidatorBLS) (*types.MsgCreateValidatorBLSResponse, error) {
	panic("TODO: unimplemented")
}

func (k msgServer) WrappedDelegate(goCtx context.Context, msg *types.MsgWrappedDelegate) (*types.MsgWrappedDelegateResponse, error) {
	panic("TODO: unimplemented")
}

func (k msgServer) WrappedUndelegate(goCtx context.Context, msg *types.MsgWrappedUndelegate) (*types.MsgWrappedUndelegateResponse, error) {
	panic("TODO: unimplemented")
}

func (k msgServer) WrappedBeginRedelegate(goCtx context.Context, msg *types.MsgWrappedBeginRedelegate) (*types.MsgWrappedBeginRedelegateResponse, error) {
	panic("TODO: unimplemented")
}
