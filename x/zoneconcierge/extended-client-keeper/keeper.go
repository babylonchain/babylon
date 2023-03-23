package extended_client_keeper

import (
	sdkerrors "cosmossdk.io/errors"
	metrics "github.com/armon/go-metrics"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

// ExtendedKeeper is same as the original clientkeeper.Keeper, except that
//   - it provides hooks for notifying other modules on received headers
//   - it applies different verification rules on received headers
//     (notably, intercepting headers rather than freezing clients upon errors that indicate dishonest majority)
type ExtendedKeeper struct {
	clientkeeper.Keeper
	cdc   codec.BinaryCodec // since some code needs to use k.cdc
	hooks ClientHooks
}

// GetHeaderInfo returns the information necessary for header timestamping or nil
// if provided message is not a header
func GetHeaderInfo(ctx sdk.Context, m exported.ClientMessage) *HeaderInfo {
	switch msg := m.(type) {
	case *ibctmtypes.Header:
		return &HeaderInfo{
			Hash:     msg.Header.LastCommitHash,
			ChaindId: msg.Header.ChainID,
			Height:   uint64(msg.Header.Height),
		}
	default:
		return nil
	}
}

// NewExtendedKeeper creates a new NewExtendedKeeper instance
func NewExtendedKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace, sk types.StakingKeeper, uk types.UpgradeKeeper) ExtendedKeeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	k := clientkeeper.NewKeeper(cdc, key, paramSpace, sk, uk)
	return ExtendedKeeper{
		Keeper: k,
		cdc:    cdc,
		hooks:  nil,
	}
}

// SetHooks sets the hooks for ExtendedKeeper
func (ek *ExtendedKeeper) SetHooks(ch ClientHooks) *ExtendedKeeper {
	if ek.hooks != nil {
		panic("cannot set hooks twice")
	}
	ek.hooks = ch

	return ek
}

// UpdateClient updates the consensus state and the state root from a provided header.
// The implementation is the same as the original IBC-Go implementation, except from:
// 1. Not freezing the client when finding a misbehaviour for header message
// 2. Calling a AfterHeaderWithValidCommit callback when receiving valid header messages (either misbehaving or not)
func (k ExtendedKeeper) UpdateClient(ctx sdk.Context, clientID string, clientMsg exported.ClientMessage) error {
	// In case of nil message nothing changes in comparison to the original IBC-Go implementation
	if clientMsg == nil {
		return k.Keeper.UpdateClient(ctx, clientID, clientMsg)
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "cannot update client with ID %s", clientID)
	}

	clientStore := k.ClientStore(ctx, clientID)

	if status := clientState.Status(ctx, clientStore, k.cdc); status != exported.Active {
		return sdkerrors.Wrapf(types.ErrClientNotActive, "cannot update client (%s) with status %s", clientID, status)
	}

	if err := clientState.VerifyClientMessage(ctx, k.cdc, clientStore, clientMsg); err != nil {
		return err
	}

	foundMisbehaviour := clientState.CheckForMisbehaviour(ctx, k.cdc, clientStore, clientMsg)

	headerInfo := GetHeaderInfo(ctx, clientMsg)

	// found misbehaviour and it was not an header, freeze client
	if foundMisbehaviour && headerInfo == nil {
		clientState.UpdateStateOnMisbehaviour(ctx, k.cdc, clientStore, clientMsg)

		k.Logger(ctx).Info("client frozen due to misbehaviour", "client-id", clientID)

		defer telemetry.IncrCounterWithLabels(
			[]string{"ibc", "client", "misbehaviour"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.LabelClientType, clientState.ClientType()),
				telemetry.NewLabel(types.LabelClientID, clientID),
				telemetry.NewLabel(types.LabelMsgType, "update"),
			},
		)

		clientkeeper.EmitSubmitMisbehaviourEvent(ctx, clientID, clientState)

		return nil
	} else if foundMisbehaviour && headerInfo != nil {
		// found misbehaviour and it was an header, this is most probably means
		// conflicting headers misbehaviour.
		ctx.Logger().Debug("received a header that has QC but is on a fork")
		txHash := tmhash.Sum(ctx.TxBytes())
		k.AfterHeaderWithValidCommit(ctx, txHash, headerInfo, true)
		return nil
	}

	// there was no misbehaviour and we receivied an header, call the callback
	if headerInfo != nil {
		txHash := tmhash.Sum(ctx.TxBytes()) // get hash of the tx that includes this header
		k.AfterHeaderWithValidCommit(ctx, txHash, headerInfo, false)
	}

	consensusHeights := clientState.UpdateState(ctx, k.cdc, clientStore, clientMsg)

	k.Logger(ctx).Info("client state updated", "client-id", clientID, "heights", consensusHeights)

	defer telemetry.IncrCounterWithLabels(
		[]string{"ibc", "client", "update"},
		1,
		[]metrics.Label{
			telemetry.NewLabel(types.LabelClientType, clientState.ClientType()),
			telemetry.NewLabel(types.LabelClientID, clientID),
			telemetry.NewLabel(types.LabelUpdateType, "msg"),
		},
	)

	// emitting events in the keeper emits for both begin block and handler client updates
	clientkeeper.EmitUpdateClientEvent(ctx, clientID, clientState.ClientType(), consensusHeights, k.cdc, clientMsg)

	return nil
}
