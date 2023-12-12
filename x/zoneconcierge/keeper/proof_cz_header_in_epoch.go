package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: fuzz tests for verifying proofs generated from KVStore

func getCZHeaderKey(chainID string, height uint64) []byte {
	key := types.CanonicalChainKey
	key = append(key, []byte(chainID)...)
	key = append(key, sdk.Uint64ToBigEndian(height)...)
	return key
}

func (k Keeper) ProveCZHeaderInEpoch(_ context.Context, header *types.IndexedHeader, epoch *epochingtypes.Epoch) (*tmcrypto.ProofOps, error) {
	czHeaderKey := getCZHeaderKey(header.ChainId, header.Height)
	_, _, proof, err := k.QueryStore(types.StoreKey, czHeaderKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func VerifyCZHeaderInEpoch(header *types.IndexedHeader, epoch *epochingtypes.Epoch, proof *tmcrypto.ProofOps) error {
	// nil check
	if header == nil {
		return fmt.Errorf("header is nil")
	} else if epoch == nil {
		return fmt.Errorf("epoch is nil")
	} else if proof == nil {
		return fmt.Errorf("proof is nil")
	}

	// sanity check
	if err := header.ValidateBasic(); err != nil {
		return err
	} else if err := epoch.ValidateBasic(); err != nil {
		return err
	}

	// ensure epoch number is same in epoch and CZ header
	if epoch.EpochNumber != header.BabylonEpoch {
		return fmt.Errorf("epoch.EpochNumber (%d) is not equal to header.BabylonEpoch (%d)", epoch.EpochNumber, header.BabylonEpoch)
	}

	// get the Merkle root, i.e., the AppHash of the sealer header
	root := epoch.SealerHeaderHash

	// Ensure The header is committed to the AppHash of the sealer header
	headerBytes, err := header.Marshal()
	if err != nil {
		return err
	}
	if err := VerifyStore(root, types.StoreKey, getCZHeaderKey(header.ChainId, header.Height), headerBytes, proof); err != nil {
		return errorsmod.Wrapf(types.ErrInvalidMerkleProof, "invalid inclusion proof for CZ header: %v", err)
	}

	return nil
}
