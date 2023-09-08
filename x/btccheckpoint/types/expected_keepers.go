package types

import (
	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"

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
	BlockHeight(ctx sdk.Context, headerHash *bbn.BTCHeaderHashBytes) (uint64, error)

	// MainChainDepth returns the depth of the header in the main chain or -1 if it does not exist in it
	// Error is returned if header is unknown to lightclient
	MainChainDepth(ctx sdk.Context, headerBytes *bbn.BTCHeaderHashBytes) (int64, error)
}

type CheckpointingKeeper interface {
	VerifyCheckpoint(ctx sdk.Context, checkpoint txformat.RawBtcCheckpoint) error
	// It quite mouthfull to have 4 different methods to operate on checkpoint state
	// but this approach decouples both modules a bit more than having some kind
	// of shared enum passed into the methods. Both modules are free to evolve their
	// representation of checkpoint state independently

	// SetCheckpointSubmitted informs checkpointing module that checkpoint was
	// successfully submitted on btc chain.
	SetCheckpointSubmitted(ctx sdk.Context, epoch uint64)
	// SetCheckpointConfirmed informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least K-deep on the main chain
	SetCheckpointConfirmed(ctx sdk.Context, epoch uint64)
	// SetCheckpointFinalized informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least W-deep on the main chain
	SetCheckpointFinalized(ctx sdk.Context, epoch uint64)

	// SetCheckpointForgotten informs checkpointing module that this checkpoint lost
	// all submissions on btc chain
	SetCheckpointForgotten(ctx sdk.Context, epoch uint64)
}

type IncentiveKeeper interface {
	RewardBTCTimestamping(ctx sdk.Context, epoch uint64, rewardDistInfo *RewardDistInfo)
}
