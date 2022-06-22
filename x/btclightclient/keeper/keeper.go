package keeper

import (
	"fmt"
	"github.com/btcsuite/btcd/wire"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// InsertHeader inserts a btcd header into the header state
func (k Keeper) InsertHeader(ctx sdk.Context, header *wire.BlockHeader) error {
	headerHash := header.BlockHash()
	if k.HeadersState(ctx).HeaderExists(&headerHash) {
		return types.ErrDuplicateHeader.Wrap("header with provided hash already exists")
	}

	if !k.HeadersState(ctx).HeaderExists(&header.PrevBlock) {
		return types.ErrHeaderParentDoesNotExist.Wrap("parent for provided hash is not maintained")
	}

	height, err := k.HeadersState(ctx).GetHeaderHeight(&header.PrevBlock)
	if err != nil {
		// Parent should always exist
		panic("Height for parent is not maintained")
	}

	k.HeadersState(ctx).CreateHeader(header, height+1)
	return nil
}
