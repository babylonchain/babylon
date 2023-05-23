package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		ics4Wrapper         types.ICS4Wrapper
		channelKeeper       types.ChannelKeeper
		portKeeper          types.PortKeeper
		authKeeper          types.AccountKeeper
		bankKeeper          types.BankKeeper
		btclcKeeper         types.BTCLightClientKeeper
		checkpointingKeeper types.CheckpointingKeeper
		btccKeeper          types.BtcCheckpointKeeper
		epochingKeeper      types.EpochingKeeper
		tmClient            types.TMClient
		storeQuerier        sdk.Queryable
		scopedKeeper        types.ScopedKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ics4Wrapper types.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	btclcKeeper types.BTCLightClientKeeper,
	checkpointingKeeper types.CheckpointingKeeper,
	btccKeeper types.BtcCheckpointKeeper,
	epochingKeeper types.EpochingKeeper,
	tmClient types.TMClient,
	storeQuerier sdk.Queryable,
	scopedKeeper types.ScopedKeeper,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		ics4Wrapper:         ics4Wrapper,
		channelKeeper:       channelKeeper,
		portKeeper:          portKeeper,
		authKeeper:          authKeeper,
		bankKeeper:          bankKeeper,
		btclcKeeper:         btclcKeeper,
		checkpointingKeeper: checkpointingKeeper,
		btccKeeper:          btccKeeper,
		epochingKeeper:      epochingKeeper,
		tmClient:            tmClient,
		storeQuerier:        storeQuerier,
		scopedKeeper:        scopedKeeper,
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
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the transfer module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
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
