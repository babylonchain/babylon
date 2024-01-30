package keeper

import (
	"testing"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	bapp "github.com/babylonchain/babylon/app"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

func BTCLightClientKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	return BTCLightClientKeeperWithCustomParams(t, types.DefaultParams())
}

func BTCLightClientKeeperWithCustomParams(t testing.TB, p types.Params) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(t), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	testCfg := bbn.ParseBtcOptionsFromConfig(bapp.EmptyAppOptions{})

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		testCfg,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())
	ctx = ctx.WithHeaderInfo(header.Info{})

	if err := k.SetParams(ctx, p); err != nil {
		panic(err)
	}

	return &k, ctx
}
