package keeper

import (
	"testing"

	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
)

func NewBTCCheckpointKeeper(
	t testing.TB,
	lk btcctypes.BTCLightClientKeeper,
	ek btcctypes.CheckpointingKeeper,
	powLimit *big.Int) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(btcctypes.StoreKey)
	tstoreKey := sdk.NewTransientStoreKey(btcctypes.TStoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(btcctypes.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		btcctypes.Amino,
		storeKey,
		memStoreKey,
		"BTCCheckpointParams",
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		tstoreKey,
		memStoreKey,
		paramsSubspace,
		lk,
		ek,
		powLimit,
		// use MainTag tests
		txformat.MainTag(0),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, btcctypes.DefaultParams())

	return &k, ctx
}
