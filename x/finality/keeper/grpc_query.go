package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
)

var _ types.QueryServer = Keeper{}

// ListPublicRandomness returns a list of public randomness committed by a given
// BTC validator
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

func (k Keeper) Block(ctx context.Context, req *types.QueryBlockRequest) (*types.QueryBlockResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	b, err := k.GetBlock(sdkCtx, req.Height)
	if err != nil {
		return nil, err
	}

	return &types.QueryBlockResponse{Block: b}, nil
}

// ListBlocks returns a list of blocks at the given finalisation status
func (k Keeper) ListBlocks(ctx context.Context, req *types.QueryListBlocksRequest) (*types.QueryListBlocksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.blockStore(sdkCtx)
	var ibs []*types.IndexedBlock
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var ib types.IndexedBlock
		k.cdc.MustUnmarshal(value, &ib)

		// hit if the queried status matches the block status, or the querier wants blocks in any state
		if (req.Status == types.QueriedBlockStatus_FINALIZED && ib.Finalized) ||
			(req.Status == types.QueriedBlockStatus_NON_FINALIZED && !ib.Finalized) ||
			(req.Status == types.QueriedBlockStatus_ANY) {
			if accumulate {
				ibs = append(ibs, &ib)
			}
			return true, nil
		}

		return false, nil
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

// VotesAtHeight returns the set of votes at a given Babylon height
func (k Keeper) VotesAtHeight(ctx context.Context, req *types.QueryVotesAtHeightRequest) (*types.QueryVotesAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// get the sig set of babylon block at given height
	btcPks := []bbn.BIP340PubKey{}
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

// Evidence returns the first evidence that allows to extract the BTC validator's SK
// associated with the given BTC validator's PK.
func (k Keeper) Evidence(ctx context.Context, req *types.QueryEvidenceRequest) (*types.QueryEvidenceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valBTCPK, err := bbn.NewBIP340PubKeyFromHex(req.ValBtcPkHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal validator BTC PK hex: %v", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	evidence := k.GetFirstSlashableEvidence(sdkCtx, valBTCPK)
	if evidence == nil {
		return nil, types.ErrNoSlashableEvidence
	}

	resp := &types.QueryEvidenceResponse{
		Evidence: evidence,
	}
	return resp, nil
}
