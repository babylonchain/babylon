package extended_client_keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ClientHooks defines the hook interface for client
type ClientHooks interface {
	AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *HeaderInfo, isOnFork bool)
}

type HeaderInfo struct {
	Hash     []byte
	ChaindId string
	Height   uint64
	Time     time.Time
}

// MultiClientHooks is a concrete implementation of ClientHooks
// It allows other modules to hook onto client ExtendedKeeper
var _ ClientHooks = &MultiClientHooks{}

type MultiClientHooks []ClientHooks

func NewMultiClientHooks(hooks ...ClientHooks) MultiClientHooks {
	return hooks
}

// invoke hooks in each keeper that hooks onto ExtendedKeeper
func (h MultiClientHooks) AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *HeaderInfo, isOnFork bool) {
	for i := range h {
		h[i].AfterHeaderWithValidCommit(ctx, txHash, header, isOnFork)
	}
}

// ensure ExtendedKeeper implements ClientHooks interfaces
var _ ClientHooks = ExtendedKeeper{}

func (ek ExtendedKeeper) AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *HeaderInfo, isOnFork bool) {
	if ek.hooks != nil {
		ek.hooks.AfterHeaderWithValidCommit(ctx, txHash, header, isOnFork)
	}
}
