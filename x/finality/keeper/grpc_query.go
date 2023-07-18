package keeper

import (
	"context"
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ListPublicRandomness(ctx context.Context, req *types.QueryListPublicRandomnessRequest) (*types.QueryListPublicRandomnessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valBTCPK, err := bbn.NewBIP340PubKeyFromHex(req.ValBtcPkHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal validator BTC PK hex: %v", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.pubRandStore(sdkCtx, valBTCPK)
	pubRandMap := map[uint64]*bbn.SchnorrPubRand{}
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		height := sdk.BigEndianToUint64(key)
		pubRand, err := bbn.NewSchnorrPubRand(value)
		if err != nil {
			panic("failed to unmarshal EOTS public randomness in KVStore")
		}
		pubRandMap[height] = pubRand
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryListPublicRandomnessResponse{
		PubRandMap: pubRandMap,
		Pagination: pageRes,
	}
	return resp, nil
}

func (k Keeper) ListBlocks(ctx context.Context, req *types.QueryListBlocksRequest) (*types.QueryListBlocksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.blockStore(sdkCtx)
	ibs := []*types.IndexedBlock{}
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var ib types.IndexedBlock
		k.cdc.MustUnmarshal(value, &ib)
		if ib.Finalized == req.Finalized {
			ibs = append(ibs, &ib)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryListBlocksResponse{
		Blocks:     ibs,
		Pagination: pageRes,
	}
	return resp, nil
}

func (k Keeper) VotesAtHeight(ctx context.Context, req *types.QueryVotesAtHeightRequest) (*types.QueryVotesAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var btcPks []bbn.BIP340PubKey

	// get the sig set of babylon block at given height
	sigSet := k.GetSigSet(sdkCtx, req.Height)
	for pkHex := range sigSet {
		pk, err := bbn.NewBIP340PubKeyFromHex(pkHex)
		if err != nil {
			// failing to unmarshal validator BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}

		btcPks = append(btcPks, pk.MustMarshal())
	}

	return &types.QueryVotesAtHeightResponse{BtcPks: btcPks}, nil
}
