package keeper

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/crypto/eots"
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

// AddFinalitySig adds a new vote to a given block
func (ms msgServer) AddFinalitySig(goCtx context.Context, req *types.MsgAddFinalitySig) (*types.MsgAddFinalitySigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure the BTC validator has voting power at this height
	valPK := req.ValBtcPk
	if ms.BTCStakingKeeper.GetVotingPower(ctx, valPK.MustMarshal(), req.BlockHeight) == 0 {
		return nil, fmt.Errorf("the BTC validator %v does not have voting power at height %d", valPK.MustMarshal(), req.BlockHeight)
	}

	// ensure the BTC validator has not casted the same vote yet
	existingSig, err := ms.GetSig(ctx, req.BlockHeight, valPK)
	if err == nil && existingSig.Equals(req.FinalitySig) {
		return nil, fmt.Errorf("the BTC validator %v has casted the same vote before", valPK.MustMarshal())
	}

	// ensure the BTC validator has committed public randomness
	pubRand, err := ms.GetPubRand(ctx, valPK, req.BlockHeight)
	if err != nil {
		return nil, err
	}

	// verify EOTS signature w.r.t. public randomness
	valBTCPK, err := valPK.ToBTCPK()
	if err != nil {
		return nil, err
	}
	if err := eots.Verify(valBTCPK, pubRand.ToFieldVal(), req.MsgToSign(), req.FinalitySig.ToModNScalar()); err != nil {
		return nil, err
	}

	// verify whether the voted block is a fork or not
	indexedBlock, err := ms.GetBlock(ctx, req.BlockHeight)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(indexedBlock.LastCommitHash, req.BlockLastCommitHash) {
		// the BTC validator votes for a fork!
		sig2, err := ms.GetSig(ctx, req.BlockHeight, valPK)
		if err != nil {
			return nil, fmt.Errorf("the BTC validator %v votes for a fork, but does not vote for the canonical block", valPK.MustMarshal())
		}
		// the BTC validator votes for a fork AND the canonical block
		// slash it via extracting its secret key
		btcSK, err := eots.Extract(valBTCPK, pubRand.ToFieldVal(), req.MsgToSign(), req.FinalitySig.ToModNScalar(), indexedBlock.MsgToSign(), sig2.ToModNScalar())
		if err != nil {
			panic(fmt.Errorf("failed to extract secret key from two EOTS signatures with the same public randomness: %v", err))
		}
		return nil, fmt.Errorf("the BTC validator %v votes two conflicting blocks! extracted secret key: %v", valPK.MustMarshal(), btcSK.Serialize())
		// TODO: what to do with the extracted secret key? e.g., have a KVStore that stores extracted SKs/forked blocks
	}
	// TODO: it's also possible that the validator votes for a fork first, then vote for canonical
	// block. We need to save the signatures on the fork, and add a detection here

	// all good, add vote to DB
	ms.SetSig(ctx, req.BlockHeight, valPK, req.FinalitySig)
	return &types.MsgAddFinalitySigResponse{}, nil
}

// CommitPubRandList commits a list of EOTS public randomness
func (ms msgServer) CommitPubRandList(goCtx context.Context, req *types.MsgCommitPubRandList) (*types.MsgCommitPubRandListResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure the request contains enough number of public randomness
	minPubRand := ms.GetParams(ctx).MinPubRand
	givenPubRand := len(req.PubRandList)
	if uint64(givenPubRand) < minPubRand {
		return nil, fmt.Errorf("the request contains too few public randomness (required minimum: %d, actual: %d)", minPubRand, givenPubRand)
	}

	// ensure the BTC validator is registered
	valBTCPKBytes := req.ValBtcPk.MustMarshal()
	if !ms.BTCStakingKeeper.HasBTCValidator(ctx, valBTCPKBytes) {
		return nil, fmt.Errorf("the validator with BTC PK %v is not registered", valBTCPKBytes)
	}

	// this BTC validator has not commit any public randomness,
	// commit the given public randomness list and return
	if ms.IsFirstPubRand(ctx, req.ValBtcPk) {
		ms.setPubRandList(ctx, req.ValBtcPk, req.StartHeight, req.PubRandList)
		return &types.MsgCommitPubRandListResponse{}, nil
	}

	// ensure height and req.StartHeight do not overlap, i.e., height < req.StartHeight
	height, _, err := ms.GetLastPubRand(ctx, req.ValBtcPk)
	if err != nil {
		return nil, err
	}
	if height >= req.StartHeight {
		return nil, fmt.Errorf("the start height (%d) has overlap with the height of the highest public randomness (%d)", req.StartHeight, height)
	}

	// all good, commit the given public randomness list
	ms.setPubRandList(ctx, req.ValBtcPk, req.StartHeight, req.PubRandList)
	return &types.MsgCommitPubRandListResponse{}, nil
}
