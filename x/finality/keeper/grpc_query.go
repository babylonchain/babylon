package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"

	bbn "github.com/babylonchain/babylon/types"

	"google.golang.org/grpc/status"

	"github.com/babylonchain/babylon/x/finality/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) VotesAtHeight(ctx context.Context, req *types.QueryVotesAtHeightRequest) (*types.QueryVotesAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var btcPks []bbn.BIP340PubKey

	// get the validator set of block at given height
	valSet := k.BTCStakingKeeper.GetVotingPowerTable(sdkCtx, req.Height)
	for pkHex := range valSet {
		pk, err := bbn.NewBIP340PubKeyFromHex(pkHex)
		if err != nil {
			// failing to unmarshal validator BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}

		btcPks = append(btcPks, pk.MustMarshal())
	}

	return &types.QueryVotesAtHeightResponse{BtcPks: btcPks}, nil
}
