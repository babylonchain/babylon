package keeper

import (
	"context"
	"math/rand"
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
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

func BTCLightClientKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx, _ := BTCLightClientKeeperWithCustomParams(t, types.DefaultParams())
	return k, ctx
}

func ChainToChainBytes(chain []*wire.BlockHeader) []bbn.BTCHeaderBytes {
	chainBytes := make([]bbn.BTCHeaderBytes, len(chain))
	for i, header := range chain {
		chainBytes[i] = bbn.NewBTCHeaderBytesFromBlockHeader(header)
	}
	return chainBytes
}

// this function must not be used at difficulty adjustment boundaries, as then
// difficulty adjustment calculation will fail
func BTCLightGenRandomChain(
	t *testing.T,
	r *rand.Rand,
	k *keeper.Keeper,
	ctx context.Context,
	initialHeight uint64,
	chainLength uint64,
) (*types.BTCHeaderInfo, *datagen.BTCHeaderPartialChain) {
	genesisHeader := datagen.NewBTCHeaderChainWithLength(r, initialHeight, 0, 1)
	genesisHeaderInfo := genesisHeader.GetChainInfo()[0]
	k.SetBaseBTCHeader(ctx, *genesisHeaderInfo)
	randomChain := datagen.NewBTCHeaderChainFromParentInfo(
		r,
		genesisHeaderInfo,
		uint32(chainLength),
	)
	err := k.InsertHeaders(ctx, randomChain.ChainToBytes())
	require.NoError(t, err)
	tip := k.GetTipInfo(ctx)
	randomChainTipInfo := randomChain.GetTipInfo()
	require.True(t, tip.Eq(randomChainTipInfo))
	return genesisHeaderInfo, randomChain
}

func BTCLightClientKeeperWithCustomParams(t testing.TB, p types.Params) (*keeper.Keeper, sdk.Context, corestore.KVStoreService) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(t), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	testCfg := bbn.ParseBtcOptionsFromConfig(bapp.EmptyAppOptions{})

	stServ := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(
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
