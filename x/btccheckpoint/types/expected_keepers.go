package types

import (
	btypes "github.com/babylonchain/babylon/types"
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
	// part of known chain. In case this is true it shoudld return this block height
	// in case this is false it should return error
	BlockHeight(headerHash btypes.BTCHeaderHashBytes) (uint64, error)

	// IsAncestor should check if childHash header is direct ancestor of parentHash
	// if either of this header is not known to light clinet it should return error
	IsAncestor(parentHash btypes.BTCHeaderHashBytes, childHash btypes.BTCHeaderHashBytes) (bool, error)
}

type CheckpointingKeeper interface {
	// CheckpointEpoch should return epoch index if provided rawCheckpoint
	// passes all checkpointing validations and error otherwise
	CheckpointEpoch(rawCheckpoint []byte) (uint64, error)
}
