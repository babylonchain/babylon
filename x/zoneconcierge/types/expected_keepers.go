package types

import (
	"context"

	bbn "github.com/babylonchain/babylon/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
}

// ICS4Wrapper defines the expected ICS4Wrapper for middleware
type ICS4Wrapper interface {
	SendPacket(
		ctx sdk.Context,
		channelCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (uint64, error)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetAllChannels(ctx sdk.Context) (channels []channeltypes.IdentifiedChannel)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, ibcexported.ClientState, error)
}

// ClientKeeper defines the expected IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (ibcexported.ClientState, bool)
	SetClientState(ctx sdk.Context, clientID string, clientState ibcexported.ClientState)
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connection connectiontypes.ConnectionEnd, found bool)
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// ScopedKeeper defines the expected x/capability scoped keeper interface
type ScopedKeeper interface {
	GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool)
	AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool
	LookupModules(ctx sdk.Context, name string) ([]string, *capabilitytypes.Capability, error)
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
}

type BTCLightClientKeeper interface {
	GetTipInfo(ctx context.Context) *btclctypes.BTCHeaderInfo
	GetMainChainFrom(ctx context.Context, startHeight uint64) []*btclctypes.BTCHeaderInfo
	GetMainChainUpTo(ctx context.Context, depth uint64) []*btclctypes.BTCHeaderInfo
	GetHeaderByHash(ctx context.Context, hash *bbn.BTCHeaderHashBytes) *btclctypes.BTCHeaderInfo
}

type BtcCheckpointKeeper interface {
	GetParams(ctx context.Context) (p btcctypes.Params)
	GetEpochData(ctx context.Context, e uint64) *btcctypes.EpochData
	GetBestSubmission(ctx context.Context, e uint64) (btcctypes.BtcStatus, *btcctypes.SubmissionKey, error)
	GetSubmissionData(ctx context.Context, sk btcctypes.SubmissionKey) *btcctypes.SubmissionData
	GetEpochBestSubmissionBtcInfo(ctx context.Context, ed *btcctypes.EpochData) *btcctypes.SubmissionBtcInfo
}

type CheckpointingKeeper interface {
	GetBLSPubKeySet(ctx context.Context, epochNumber uint64) ([]*checkpointingtypes.ValidatorWithBlsKey, error)
	GetRawCheckpoint(ctx context.Context, epochNumber uint64) (*checkpointingtypes.RawCheckpointWithMeta, error)
	GetFinalizedEpoch(ctx context.Context) (uint64, error)
}

type EpochingKeeper interface {
	GetHistoricalEpoch(ctx context.Context, epochNumber uint64) (*epochingtypes.Epoch, error)
	GetEpoch(ctx context.Context) *epochingtypes.Epoch
}

// CometClient is a Comet client that allows to query tx inclusion proofs
type CometClient interface {
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
}
