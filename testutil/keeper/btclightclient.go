package keeper

import (
	"testing"

	"cosmossdk.io/core/header"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/btcsuite/btcd/wire"
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
	btclightclientk "github.com/babylonchain/babylon/x/btclightclient/keeper"
	btclightclientt "github.com/babylonchain/babylon/x/btclightclient/types"
)

func BTCLightClientKeeper(t testing.TB) (*btclightclientk.Keeper, sdk.Context) {
	k, ctx, _ := BTCLightClientKeeperWithCustomParams(t, btclightclientt.DefaultParams())
	return k, ctx
}

// NewBTCHeaderBytesList takes a list of block headers and parses it to BTCHeaderBytes.
func NewBTCHeaderBytesList(chain []*wire.BlockHeader) []bbn.BTCHeaderBytes {
	chainBytes := make([]bbn.BTCHeaderBytes, len(chain))
	for i, header := range chain {
		chainBytes[i] = bbn.NewBTCHeaderBytesFromBlockHeader(header)
	}
	return chainBytes
}

func BTCLightClientKeeperWithCustomParams(t testing.TB, p btclightclientt.Params) (*btclightclientk.Keeper, sdk.Context, corestore.KVStoreService) {
	storeKey := storetypes.NewKVStoreKey(btclightclientt.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(t), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	testCfg := bbn.ParseBtcOptionsFromConfig(bapp.EmptyAppOptions{})

	stServ := runtime.NewKVStoreService(storeKey)
	k := btclightclientk.NewKeeper(
		cdc,
		stServ,
		testCfg,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())
	ctx = ctx.WithHeaderInfo(header.Info{})

	if err := k.SetParams(ctx, p); err != nil {
		panic(err)
	}

	return &k, ctx, stServ
}
