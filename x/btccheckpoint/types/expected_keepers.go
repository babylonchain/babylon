package types

import (
	bbl "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type BTCLightClientKeeper interface {
	// BlockHeight should validate if header with given hash is valid and if it is
	// part of known chain. In case this is true it should return this block height
	// in case this is false it should return error
	BlockHeight(ctx sdk.Context, headerHash *bbl.BTCHeaderHashBytes) (uint64, error)

	// IsAncestor should check if childHash header is direct ancestor of parentHash
	// if either of this header is not known to light clinet it should return error
	IsAncestor(ctx sdk.Context, parentHash *bbl.BTCHeaderHashBytes, childHash *bbl.BTCHeaderHashBytes) (bool, error)

	// MainChainDepth returns the depth of the header in the main chain or -1 if it does not exist in it
	// Error is returned if header is unknown to lightclient
	MainChainDepth(ctx sdk.Context, headerBytes *bbl.BTCHeaderHashBytes) (int64, error)
}

type CheckpointingKeeper interface {
	// CheckpointEpoch should return epoch index if provided rawCheckpoint
	// passes all checkpointing validations and error otherwise
	CheckpointEpoch(ctx sdk.Context, rawCheckpoint []byte) (uint64, error)

	// It quite mouthfull to have 4 different methods to operate on checkpoint state
	// but this approach decouples both modules a bit more than having some kind
	// of shared enum passed into the methods. Both modules are free to evolve their
	// representation of checkpoint state independently

	// SetCheckpointSubmitted informs checkpointing module that checkpoint was
	// successfully submitted on btc chain. It can be either or main chain or fork.
	SetCheckpointSubmitted(ctx sdk.Context, rawCheckpoint []byte) error
	// SetCheckpointConfirmed informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least K-deep on the main chain
	SetCheckpointConfirmed(ctx sdk.Context, rawCheckpoint []byte) error
	// SetCheckpointFinalized informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least W-deep on the main chain
	SetCheckpointFinalized(ctx sdk.Context, rawCheckpoint []byte) error

	// SetCheckpointForgotten informs checkpointing module that this checkpoint lost
	// all submissions on btc chain
	SetCheckpointForgotten(ctx sdk.Context, rawCheckpoint []byte) error
}
