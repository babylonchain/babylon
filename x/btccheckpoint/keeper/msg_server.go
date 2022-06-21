package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

func (m msgServer) InsertBTCSpvProof(ctx context.Context, req *types.MsgInsertBTCSpvProof) (*types.MsgInsertBTCSpvProofResponse, error) {
	//TODO implement me
	panic("implement me")
}

var _ types.MsgServer = msgServer{}
