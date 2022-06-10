package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

type msgServer struct {
    // This should be a reference to Keeper 
	k Keeper
}

func (m msgServer) InsertHeader(ctx context.Context, header *types.MsgInsertHeader) (*types.MsgInsertHeaderResponse, error) {
	//TODO implement me
	panic("implement me")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}
