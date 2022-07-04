package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/checkpointing/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) AddBlsSig(ctx context.Context, header *types.MsgAddBlsSig) (*types.MsgAddBlsSigResponse, error) {
	panic("TODO: implement me")
}

func (m msgServer) CreateBlsKey(ctx context.Context, msg *types.MsgCreateBlsKey) (*types.MsgCreateBlsKeyResponse, error) {
	panic("TODO: implement me")
}

func (m msgServer) WrappedCreateValidator(ctx context.Context, msg *types.MsgWrappedCreateValidator) (*types.MsgWrappedCreateValidatorResponse, error) {
	panic("TODO: implement me")
}
