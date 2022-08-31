package keeper

import (
	"testing"

	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func NewBTCCheckpointKeeper(
	t testing.TB,
	lk btcctypes.BTCLightClientKeeper,
	ek btcctypes.CheckpointingKeeper,
	kDeep uint64,
	wDeep uint64,
	powLimit *big.Int) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(btcctypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(btcctypes.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
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
		memStoreKey,
		paramsSubspace,
		lk,
		ek,
		kDeep,
		wDeep,
		powLimit,
		// use MainTag tests
		txformat.MainTag,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, btcctypes.DefaultParams())

	return &k, ctx
}
