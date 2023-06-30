package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// AddVote adds a new vote to a given block
func (ms msgServer) AddVote(goCtx context.Context, req *types.MsgAddVote) (*types.MsgAddVoteResponse, error) {
	panic("TODO: implement me")
}

// CommitPubRand commits a list of EOTS public randomness
func (ms msgServer) CommitPubRand(goCtx context.Context, req *types.MsgCommitPubRand) (*types.MsgCommitPubRandResponse, error) {
	panic("TODO: implement me")
}
