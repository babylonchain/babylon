package datagen

import (
	"math/rand"

	abci "github.com/cometbft/cometbft/abci/types"

	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

func GenRandomVoteExtension(
	epochNum, height uint64,
	blockHash checkpointingtypes.BlockHash,
	valSet *GenesisValidators,
	r *rand.Rand,
) ([]abci.ExtendedVoteInfo, error) {
	genesisKeys := valSet.GetGenesisKeys()
	extendedVotes := make([]abci.ExtendedVoteInfo, 0, len(valSet.Keys))
	for i := 0; i < len(valSet.Keys); i++ {
		sig := GenRandomBlsMultiSig(r)
		ve := checkpointingtypes.VoteExtension{
			Signer:    genesisKeys[i].ValidatorAddress,
			BlockHash: blockHash,
			EpochNum:  epochNum,
			Height:    height,
			BlsSig:    &sig,
		}
		veBytes, err := ve.Marshal()
		if err != nil {
			return nil, err
		}
		veInfo := abci.ExtendedVoteInfo{VoteExtension: veBytes}
		extendedVotes = append(extendedVotes, veInfo)
	}

	return extendedVotes, nil
}
