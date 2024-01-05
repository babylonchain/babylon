package keeper

import (
	"testing"

	"cosmossdk.io/core/header"
	storemetrics "cosmossdk.io/store/metrics"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func BTCStakingKeeper(
	t testing.TB,
	btclcKeeper types.BTCLightClientKeeper,
	btccKeeper types.BtcCheckpointKeeper,
) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(t), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		btclcKeeper,
		btccKeeper,
		&chaincfg.SimNetParams,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	ctx = ctx.WithHeaderInfo(header.Info{})

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return &k, ctx
}
