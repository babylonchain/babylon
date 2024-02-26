package checkpointing

import (
	"context"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CheckpointingKeeper interface {
	GetPubKeyByConsAddr(context.Context, sdk.ConsAddress) (cmtprotocrypto.PublicKey, error)
	GetEpoch(ctx context.Context) *epochingtypes.Epoch
	GetValidatorSet(ctx context.Context, epochNumber uint64) epochingtypes.ValidatorSet
	GetTotalVotingPower(ctx context.Context, epochNumber uint64) int64
	GetBlsPubKey(ctx context.Context, address sdk.ValAddress) (bls12381.PublicKey, error)
	VerifyBLSSig(ctx context.Context, sig *types.BlsSig) error
	SealCheckpoint(ctx context.Context, ckptWithMeta *types.RawCheckpointWithMeta) error
}
