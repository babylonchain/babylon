package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
)

type msgServer struct {
	// This should be a reference to Keeper
	k Keeper
}

const BTCDifficultyMultiplier = 4 // TODO: get from config file

func (m msgServer) InsertHeader(ctx context.Context, msg *types.MsgInsertHeader) (*types.MsgInsertHeaderResponse, error) {
	// Perform the checks that checkBlockHeaderContext of btcd does
	// https://github.com/btcsuite/btcd/blob/master/blockchain/validate.go#L644
	// We skip the time, checkpoint, and version checks
	// TODO: Implement an AnteHandler that performs these checks
	// 		 so as to not pollute the mempool with transactions
	// 		 that will get rejected.
	if msg == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message is nil")
	}

	if msg.Header == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message header is nit")
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	parentHash := msg.Header.ParentHash()
	// Retrieve parent
	parent, err := m.k.HeadersState(sdkCtx).GetHeaderByHash(parentHash)
	// parent does not exist
	if err != nil {
		return nil, err
	}

	// The new block will either be the first block of a recalculation event
	// which happens every 2,016 blocks or a normal block.
	// In the second case, it's difficulty should be exactly the same as it's parent
	// while in the second case it should have a maximum difference of a factor of 4 from it
	// See: https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch10.asciidoc#retargeting-to-adjust-difficulty
	// We consolidate those into a single check.
	oldDifficulty := blockchain.CompactToBig(parent.Header.Bits())
	currentDifficulty := blockchain.CompactToBig(msg.Header.Bits())
	maxCurrentDifficulty := new(big.Int).Mul(oldDifficulty, big.NewInt(BTCDifficultyMultiplier))
	minCurrentDifficulty := new(big.Int).Div(oldDifficulty, big.NewInt(BTCDifficultyMultiplier))
	if currentDifficulty.Cmp(maxCurrentDifficulty) > 0 || currentDifficulty.Cmp(minCurrentDifficulty) < 0 {
		return nil, types.ErrInvalidDifficulty.Wrap("difficulty not relevant to parent difficulty")
	}

	// All good, insert the header
	err = m.k.InsertHeader(sdkCtx, msg.Header)
	if err != nil {
		return nil, err
	}
	return &types.MsgInsertHeaderResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}
