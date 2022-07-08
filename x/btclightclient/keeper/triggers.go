package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) triggerRollBack(ctx sdk.Context, header *wire.BlockHeader, height uint64) {
	headerHash := header.BlockHash()
	btcHeaderHashBytes := bbl.NewBTCHeaderHashBytesFromChainhash(&headerHash)
	// Trigger AfterBTCRollBack hook
	k.AfterBTCRollBack(ctx, btcHeaderHashBytes, height+1)
	// Emit BTCRollBack event
	ctx.EventManager().EmitTypedEvent(&types.EventBTCRollBack{
		Height: height + 1,
		Hash:   &btcHeaderHashBytes,
	})
}

func (k Keeper) triggerRollForward(ctx sdk.Context, header *wire.BlockHeader, height uint64) {
	headerHash := header.BlockHash()
	btcHeaderHashBytes := bbl.NewBTCHeaderHashBytesFromChainhash(&headerHash)
	// Trigger AfterBTCRollForward hook
	k.AfterBTCRollForward(ctx, btcHeaderHashBytes, height+1)
	// Emit BTCRollForward event
	ctx.EventManager().EmitTypedEvent(&types.EventBTCRollForward{
		Height: height + 1,
		Hash:   &btcHeaderHashBytes,
	})
}
