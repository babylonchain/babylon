package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		btclcKeeper types.BTCLightClientKeeper
		btccKeeper  types.BtcCheckpointKeeper

		btcNet *chaincfg.Params
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,

	btclcKeeper types.BTCLightClientKeeper,
	btccKeeper types.BtcCheckpointKeeper,

	btcNet *chaincfg.Params,
	authority string,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,

		btclcKeeper: btclcKeeper,
		btccKeeper:  btccKeeper,

		btcNet:    btcNet,
		authority: authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) getHeaderAndDepth(ctx sdk.Context, headerHash *bbn.BTCHeaderHashBytes) (*btclctypes.BTCHeaderInfo, uint64, error) {
	if headerHash == nil {
		return nil, 0, fmt.Errorf("headerHash is nil")
	}
	// get the header
	header := k.btclcKeeper.GetHeaderByHash(ctx, headerHash)
	if header == nil {
		return nil, 0, fmt.Errorf("header that includes the staking tx is not found")
	}
	// get the tip
	tip := k.btclcKeeper.GetTipInfo(ctx)
	// If the height of the requested header is larger than the tip, return -1
	if tip.Height < header.Height {
		return nil, 0, fmt.Errorf("header is higher than the tip in BTC light client")
	}
	// The depth is the number of blocks that have been build on top of the header
	// For example:
	// 		Tip: 0-deep
	// 		Tip height is 10, headerInfo height is 5: 5-deep etc.
	headerDepth := tip.Height - header.Height

	return header, headerDepth, nil
}
