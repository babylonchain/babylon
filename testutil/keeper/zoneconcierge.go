package keeper

import (
	"testing"

	"github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmdb "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/stretchr/testify/require"
)

// zoneconciergeChannelKeeper is a stub of ChannelKeeper
type zoneconciergeChannelKeeper struct{}

func (zoneconciergeChannelKeeper) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool) {
	return channeltypes.Channel{}, false
}
func (zoneconciergeChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return 0, false
}
func (zoneconciergeChannelKeeper) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	return nil
}
func (zoneconciergeChannelKeeper) ChanCloseInit(ctx sdk.Context, portID, channelID string, chanCap *capabilitytypes.Capability) error {
	return nil
}

func (zoneconciergeChannelKeeper) GetAllChannels(ctx sdk.Context) []channeltypes.IdentifiedChannel {
	return nil
}
func (zoneconciergeChannelKeeper) GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, ibcexported.ClientState, error) {
	return "", nil, nil
}

// zoneconciergeportKeeper is a stub of PortKeeper
type zoneconciergePortKeeper struct{}

func (zoneconciergePortKeeper) BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability {
	return &capabilitytypes.Capability{}
}

type zoneconciergeStoreQuerier struct{}

func (zoneconciergeStoreQuerier) Query(req abci.RequestQuery) abci.ResponseQuery {
	return abci.ResponseQuery{
		ProofOps: &tmcrypto.ProofOps{
			Ops: []tmcrypto.ProofOp{
				tmcrypto.ProofOp{},
			},
		},
	}
}

func ZoneConciergeKeeper(t testing.TB, btclcKeeper types.BTCLightClientKeeper, checkpointingKeeper types.CheckpointingKeeper, btccKeeper types.BtcCheckpointKeeper, epochingKeeper types.EpochingKeeper, tmClient types.TMClient) (*keeper.Keeper, sdk.Context) {
	logger := log.NewNopLogger()

	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(registry)
	capabilityKeeper := capabilitykeeper.NewKeeper(appCodec, storeKey, memStoreKey)
	k := keeper.NewKeeper(
		appCodec,
		storeKey,
		memStoreKey,
		nil, // TODO: mock this keeper
		nil, // TODO: mock this keeper
		zoneconciergeChannelKeeper{},
		zoneconciergePortKeeper{},
		nil, // TODO: mock this keeper
		nil, // TODO: mock this keeper
		btclcKeeper,
		checkpointingKeeper,
		btccKeeper,
		epochingKeeper,
		tmClient,
		zoneconciergeStoreQuerier{},
		capabilityKeeper.ScopeToModule("ZoneconciergeScopedKeeper"),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	return k, ctx
}
