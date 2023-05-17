package types

import (
	context "context"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
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
	GetClientConsensusState(ctx sdk.Context, clientID string) (connection ibcexported.ConsensusState, found bool)
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
	GetTipInfo(ctx sdk.Context) *btclctypes.BTCHeaderInfo
	GetBaseBTCHeader(ctx sdk.Context) *btclctypes.BTCHeaderInfo
	GetHighestCommonAncestor(ctx sdk.Context, header1 *btclctypes.BTCHeaderInfo, header2 *btclctypes.BTCHeaderInfo) *btclctypes.BTCHeaderInfo
	GetInOrderAncestorsUntil(ctx sdk.Context, descendant *btclctypes.BTCHeaderInfo, ancestor *btclctypes.BTCHeaderInfo) []*btclctypes.BTCHeaderInfo
	GetMainChainUpTo(ctx sdk.Context, depth uint64) []*btclctypes.BTCHeaderInfo
}

type BtcCheckpointKeeper interface {
	GetParams(ctx sdk.Context) (p btcctypes.Params)
	GetBestSubmission(ctx sdk.Context, e uint64) (btcctypes.BtcStatus, *btcctypes.SubmissionKey, error)
	GetSubmissionData(ctx sdk.Context, sk btcctypes.SubmissionKey) *btcctypes.SubmissionData
}

type CheckpointingKeeper interface {
	GetBLSPubKeySet(ctx sdk.Context, epochNumber uint64) ([]*checkpointingtypes.ValidatorWithBlsKey, error)
	GetRawCheckpoint(ctx sdk.Context, epochNumber uint64) (*checkpointingtypes.RawCheckpointWithMeta, error)
}

type EpochingKeeper interface {
	GetHistoricalEpoch(ctx sdk.Context, epochNumber uint64) (*epochingtypes.Epoch, error)
	GetEpoch(ctx sdk.Context) *epochingtypes.Epoch
	ProveAppHashInEpoch(ctx sdk.Context, height uint64, epochNumber uint64) (*tmcrypto.Proof, error)
}

// TMClient is a Tendermint client that allows to query tx inclusion proofs
type TMClient interface {
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
}
