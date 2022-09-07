package keeper

import (
	"context"
	"errors"
	"math"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) lowestBtcHeight(ctx sdk.Context, subKey *types.SubmissionKey) (uint64, error) {
	// initializing to max, as then every header number will be smaller
	var lowestHeaderNumber uint64 = math.MaxUint64

	for _, tk := range subKey.Key {

		if !k.CheckHeaderIsOnMainChain(ctx, tk.Hash) {
			return 0, errors.New("one of submission headers not on main chain")
		}

		headerNumber, err := k.GetBlockHeight(ctx, tk.Hash)

		if err != nil {
			// CheckHeaderIsOnMainChain  (which uses main chain depth) returned true which
			// means header should be saved and we should know its heigh but GetBlockHeight
			// returned error. Something is really bad, panic.
			panic("Inconsistent data model in btc light client")
		}

		if headerNumber < lowestHeaderNumber {
			lowestHeaderNumber = headerNumber
		}
	}

	return lowestHeaderNumber, nil
}

func (k Keeper) BtcCheckpointHeight(c context.Context, req *types.QueryBtcCheckpointHeightRequest) (*types.QueryBtcCheckpointHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	checkpointEpoch := req.GetEpochNum()

	epochData := k.GetEpochData(ctx, checkpointEpoch)

	// Check if we have any submission for given epoch
	if epochData == nil || len(epochData.Key) == 0 {
		return nil, errors.New("checkpoint for given epoch not yet submitted")
	}

	var lowestHeaderNumber uint64 = math.MaxUint64

	// we need to go for each submission in given epoch
	for _, submissionKey := range epochData.Key {

		headerNumber, err := k.lowestBtcHeight(ctx, submissionKey)

		if err != nil {
			// submission is not valid for some reason, ignore it
			continue
		}

		if headerNumber < lowestHeaderNumber {
			lowestHeaderNumber = headerNumber
		}
	}

	if lowestHeaderNumber == math.MaxUint64 {
		return nil, errors.New("there is no valid submission for given raw checkpoint")
	}

	return &types.QueryBtcCheckpointHeightResponse{EarliestBtcBlockNumber: lowestHeaderNumber}, nil
}
