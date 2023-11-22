package keeper

import (
	"context"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		ics4Wrapper         types.ICS4Wrapper
		clientKeeper        types.ClientKeeper
		channelKeeper       types.ChannelKeeper
		portKeeper          types.PortKeeper
		authKeeper          types.AccountKeeper
		bankKeeper          types.BankKeeper
		btclcKeeper         types.BTCLightClientKeeper
		checkpointingKeeper types.CheckpointingKeeper
		btccKeeper          types.BtcCheckpointKeeper
		epochingKeeper      types.EpochingKeeper
		cmtClient           types.CometClient
		storeQuerier        storetypes.Queryable
		scopedKeeper        types.ScopedKeeper
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	ics4Wrapper types.ICS4Wrapper,
	clientKeeper types.ClientKeeper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	btclcKeeper types.BTCLightClientKeeper,
	checkpointingKeeper types.CheckpointingKeeper,
	btccKeeper types.BtcCheckpointKeeper,
	epochingKeeper types.EpochingKeeper,
	cmtClient types.CometClient,
	storeQuerier storetypes.Queryable,
	scopedKeeper types.ScopedKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeService:        storeService,
		ics4Wrapper:         ics4Wrapper,
		clientKeeper:        clientKeeper,
		channelKeeper:       channelKeeper,
		portKeeper:          portKeeper,
		authKeeper:          authKeeper,
		bankKeeper:          bankKeeper,
		btclcKeeper:         btclcKeeper,
		checkpointingKeeper: checkpointingKeeper,
		btccKeeper:          btccKeeper,
		epochingKeeper:      epochingKeeper,
		cmtClient:           cmtClient,
		storeQuerier:        storeQuerier,
		scopedKeeper:        scopedKeeper,
		authority:           authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the transfer module. Used in ExportGenesis
func (k Keeper) GetPort(ctx context.Context) string {
	store := k.storeService.OpenKVStore(ctx)
	port, err := store.Get(types.PortKey)
	if err != nil {
		panic(err)
	}
	return string(port)
}

// SetPort sets the portID for the transfer module. Used in InitGenesis
func (k Keeper) SetPort(ctx context.Context, portID string) {
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(types.PortKey, []byte(portID)); err != nil {
		panic(err)
	}
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
