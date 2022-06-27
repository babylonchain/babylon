package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/btccheckpoint/btcutils"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

func (m msgServer) getProofHeight(proofs []*btcutils.ParsedProof) (uint64, error) {
	var latestblock = uint64(0)

	for _, proof := range proofs {
		// TODO consider interfaces here. As currently this implicity assume that all headers
		// are on the same fork. The only possiblity to detect that this is not the case
		// is to pass []*wire.BlockHeader to light client, and for it to make that
		// determination
		num, err := m.k.GetBlockHeight(proof.BlockHeader)

		if err != nil {
			return latestblock, err
		}

		// returning hightes block number as checkpoint number as if highest becomes
		// stable then it means older is also stable.
		if num > latestblock {
			latestblock = num
		}
	}

	return latestblock, nil
}

// returns fully assembled rawcheckpoint data and the latest header number of
// headers provided in the proof
func (m msgServer) getRawCheckPoint(proofs []*btcutils.ParsedProof) []byte {
	var rawCheckpointData []byte

	for _, proof := range proofs {
		rawCheckpointData = append(rawCheckpointData, proof.OpReturnData...)
	}

	return rawCheckpointData
}

// TODO at some point add proper logging of error
// TODO emit some events for external consumers. Those should be probably emited
// at EndBlockerCallback
func (m msgServer) InsertBTCSpvProof(ctx context.Context, req *types.MsgInsertBTCSpvProof) (*types.MsgInsertBTCSpvProofResponse, error) {
	// TODO get PowLimit from config
	proofs, e := types.ParseTwoProofs(req.Proofs, btcchaincfg.MainNetParams.PowLimit)

	if e != nil {
		return nil, types.ErrInvalidCheckpointProof
	}

	// TODO for now we do nothing with processed blockHeight but ultimatly it should
	// be a part of timestamp
	_, err := m.getProofHeight(proofs)

	if err != nil {
		return nil, err
	}

	// At this point:
	// - every proof of inclusion is valid i.e every transaction is proved to be
	// part of provided block and contains some OP_RETURN data
	// - header is proved to be part of the chain we know about thorugh BTCLightClient
	// Inform checkpointing module about it.
	rawCheckPointData := m.getRawCheckPoint(proofs)

	epochNum, err := m.k.GetCheckpointEpoch(rawCheckPointData)

	if err != nil {
		return nil, err
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// TODO consider handling here.
	// Checkpointing module deemed this checkpoint as correc so for now lets just
	// store it. (in future store some metadata about it)
	// Things to consider:
	// - Are we really guaranteed that epoch is unique key ?
	// - It would probably be better to check for duplicates ourselves, by keeping some
	// additional indexes like: sha256(checkpoint) -> epochNum and epochnNum -> checkpointStatus
	// then we could check for dupliacates without involvement of checkpointing
	// module and parsing checkpoint itself
	// - What is good db layout for all requiremens
	m.k.StoreCheckpoint(sdkCtx, epochNum, rawCheckPointData)

	return &types.MsgInsertBTCSpvProofResponse{}, nil
}

var _ types.MsgServer = msgServer{}
