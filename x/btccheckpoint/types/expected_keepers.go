package types

import (
	"github.com/btcsuite/btcd/wire"
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
	// Function should validate if provided header is valid and return header
	// height if thats the case.
	BlockHeight(header wire.BlockHeader) (uint64, error)
}

type CheckpointingKeeper interface {
	// CheckpointEpoch should return epoch index it passes all checkpoint validation
	// and error otherwise
	CheckpointEpoch(rawCheckpoint []byte) (uint64, error)
}
