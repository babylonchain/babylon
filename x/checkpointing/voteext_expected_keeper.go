package checkpointing

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

type CheckpointingKeeper interface {
	GetEpoch(ctx context.Context) *epochingtypes.Epoch
	GetBLSSignerAddress() sdk.ValAddress
	GetValidatorSet(ctx context.Context, epochNumber uint64) epochingtypes.ValidatorSet
	SignBLS(epochNum uint64, blockHash types.BlockHash) (bls12381.Signature, error)
	GetValidatorAddress() sdk.ValAddress
	VerifyBLSSig(ctx context.Context, sig *types.BlsSig) error
}
