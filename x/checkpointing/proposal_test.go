package checkpointing_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"

	"time"

	"testing"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/mocks"
	"github.com/babylonchain/babylon/x/checkpointing"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	et "github.com/babylonchain/babylon/x/epoching/types"
	cbftt "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type TestValidator struct {
	Keys  *datagen.GenesisKeyWithBLS
	Power int64
}

func (v *TestValidator) CometValidator() *cbftt.Validator {
	return &cbftt.Validator{
		Address: v.Keys.GenesisKey.ValPubkey.Address(),
		Power:   v.Power,
	}
}

func (v *TestValidator) EpochingValidator() et.Validator {
	return et.Validator{
		Addr:  v.Keys.GenesisKey.ValPubkey.Address(),
		Power: v.Power,
	}
}

func (v *TestValidator) ProtoPubkey() cmtprotocrypto.PublicKey {
	validatorPubKey := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Ed25519{
			Ed25519: v.Keys.PrivKey.PubKey().Bytes(),
		},
	}
	return validatorPubKey
}

func (v *TestValidator) VoteExtension(
	bh *checkpointingtypes.BlockHash,
	epochNum uint64,
) checkpointingtypes.VoteExtension {
	signBytes := checkpointingtypes.GetSignBytes(epochNum, *bh)
	// Need valid bls signature for aggregation
	bls := bls12381.Sign(v.Keys.PrivateKey, signBytes)

	return checkpointingtypes.VoteExtension{
		Signer:    v.Keys.ValidatorAddress,
		BlockHash: bh,
		EpochNum:  epochNum,
		Height:    0,
		BlsSig:    &bls,
	}
}

func (v *TestValidator) SignVoteExtension(
	t *testing.T,
	bytes []byte,
	height int64,
	chainId string,
) cbftt.ExtendedVoteInfo {
	votExt := genVoteExt(t,
		bytes, height, 0, chainId)
	signature, err := v.Keys.PrivKey.Sign(votExt)
	require.NoError(t, err)

	evi := cbftt.ExtendedVoteInfo{
		Validator:          *v.CometValidator(),
		VoteExtension:      bytes,
		ExtensionSignature: signature,
		BlockIdFlag:        tendermintTypes.BlockIDFlagCommit,
	}

	return evi
}

func (v *TestValidator) ValidatorAddress(t *testing.T) sdk.ValAddress {
	valAddress, err := sdk.ValAddressFromBech32(v.Keys.ValidatorAddress)
	require.NoError(t, err)
	return valAddress
}

func (v *TestValidator) BlsPubKey() bls12381.PublicKey {
	return *v.Keys.BlsKey.Pubkey
}

func genNTestValidators(t *testing.T, n int) []TestValidator {
	if n == 0 {
		return []TestValidator{}
	}

	keys, err := datagen.GenesisValidatorSet(n)
	require.NoError(t, err)

	var vals []TestValidator
	for _, key := range keys.Keys {
		k := key
		vals = append(vals, TestValidator{
			Keys:  k,
			Power: 100,
		})
	}

	return vals
}

func setupSdkCtx(height int64) sdk.Context {
	return sdk.Context{}.WithHeaderInfo(header.Info{
		Height: height,
		Time:   time.Now(),
	}).WithConsensusParams(tendermintTypes.ConsensusParams{
		Abci: &tendermintTypes.ABCIParams{
			VoteExtensionsEnableHeight: 1,
		},
	}).WithChainID("test")
}

func firstEpoch() *et.Epoch {
	return &et.Epoch{
		EpochNumber:          1,
		CurrentEpochInterval: 10,
		FirstBlockHeight:     1,
	}
}

type EpochAndCtx struct {
	Epoch *et.Epoch
	Ctx   sdk.Context
}

func epochAndVoteExtensionCtx() *EpochAndCtx {
	epoch := firstEpoch()
	ctx := setupSdkCtx(int64(epoch.FirstBlockHeight) + int64(epoch.GetCurrentEpochInterval()))
	return &EpochAndCtx{
		Epoch: epoch,
		Ctx:   ctx,
	}
}

func genVoteExt(
	t *testing.T,
	ext []byte,
	height int64,
	round int64,
	chainID string,
) []byte {
	cve := tendermintTypes.CanonicalVoteExtension{
		Extension: ext,
		Height:    height, // the vote extension was signed in the previous height
		Round:     round,
		ChainId:   chainID,
	}

	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	extSignBytes, err := marshalDelimitedFn(&cve)
	require.NoError(t, err)
	return extSignBytes
}

func requestPrepareProposal(height int64, votes []cbftt.ExtendedVoteInfo) *cbftt.RequestPrepareProposal {
	return &cbftt.RequestPrepareProposal{
		MaxTxBytes: 10000,
		Txs:        [][]byte{},
		LocalLastCommit: cbftt.ExtendedCommitInfo{
			Round: 0,
			Votes: votes,
		},
		Height: height,
	}
}

func randomBlockHash() checkpointingtypes.BlockHash {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return datagen.GenRandomBlockHash(r)
}

// TODO There should be one function to verify the checkpoint against the validator set
// but currently there are different implementations in the codebase in checpointing module
// and zonecocierge module
func verifyCheckpoint(validators []TestValidator, rawCkpt *checkpointingtypes.RawCheckpoint) error {
	valsCopy := validators

	sort.Slice(valsCopy, func(i, j int) bool {
		return sdk.BigEndianToUint64(valsCopy[i].EpochingValidator().Addr) < sdk.BigEndianToUint64(valsCopy[j].EpochingValidator().Addr)
	})

	var validatorWithBls []*checkpointingtypes.ValidatorWithBlsKey

	for _, val := range valsCopy {
		validatorWithBls = append(validatorWithBls, &checkpointingtypes.ValidatorWithBlsKey{
			ValidatorAddress: val.Keys.ValidatorAddress,
			BlsPubKey:        val.BlsPubKey(),
			VotingPower:      uint64(val.Power),
		})
	}

	valSet := &checkpointingtypes.ValidatorWithBlsKeySet{ValSet: validatorWithBls}
	// filter validator set that contributes to the signature
	signerSet, signerSetPower, err := valSet.FindSubsetWithPowerSum(rawCkpt.Bitmap)
	if err != nil {
		return err
	}
	// ensure the signerSet has > 2/3 voting power
	if signerSetPower*3 <= valSet.GetTotalPower()*2 {
		return fmt.Errorf("failed")
	}
	// verify BLS multisig
	signedMsgBytes := rawCkpt.SignedMsg()
	ok, err := bls12381.VerifyMultiSig(*rawCkpt.BlsMultiSig, signerSet.GetBLSKeySet(), signedMsgBytes)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("BLS signature does not match the public key")
	}
	return nil
}

type Scenario struct {
	TotalPower   int64
	ValidatorSet []TestValidator
	Extensions   []cbftt.ExtendedVoteInfo
}

type ValidatorsAndExtensions struct {
	Vals       []TestValidator
	Extensions []checkpointingtypes.VoteExtension
}

func generateNValidatorAndVoteExtensions(t *testing.T, n int, bh *checkpointingtypes.BlockHash, epochNumber uint64) (*ValidatorsAndExtensions, int64) {
	validators := genNTestValidators(t, n)
	var extensions []checkpointingtypes.VoteExtension
	var power int64
	for _, val := range validators {
		validator := val
		ve := validator.VoteExtension(bh, epochNumber)
		extensions = append(extensions, ve)
		power += validator.Power
	}
	return &ValidatorsAndExtensions{
		Vals:       validators,
		Extensions: extensions,
	}, power
}

func ToValidatorSet(v []TestValidator) et.ValidatorSet {
	var cv []et.Validator
	for _, val := range v {
		cv = append(cv, val.EpochingValidator())
	}
	return et.NewSortedValidatorSet(cv)
}

func TestPrepareProposalAtVoteExtensionHeight(t *testing.T) {
	tests := []struct {
		name          string
		scenarioSetup func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario
		expectError   bool
	}{
		{
			name: "Empty vote extension list ",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				validatorAndExtensions, totalPower := generateNValidatorAndVoteExtensions(t, 4, &bh, ec.Epoch.EpochNumber)
				return &Scenario{
					TotalPower:   totalPower,
					ValidatorSet: validatorAndExtensions.Vals,
					Extensions:   []cbftt.ExtendedVoteInfo{},
				}
			},
			expectError: true,
		},
		{
			name: "List with only empty vote extensions",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				validatorAndExtensions, totalPower := generateNValidatorAndVoteExtensions(t, 4, &bh, ec.Epoch.EpochNumber)
				var signedVoteExtensions []cbftt.ExtendedVoteInfo
				for i, val := range validatorAndExtensions.Vals {
					validator := val
					ek.EXPECT().GetPubKeyByConsAddr(gomock.Any(), sdk.ConsAddress(validator.ValidatorAddress(t).Bytes())).Return(validator.ProtoPubkey(), nil).AnyTimes()
					ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(nil).AnyTimes()
					ek.EXPECT().GetBlsPubKey(gomock.Any(), validator.ValidatorAddress(t)).Return(validator.BlsPubKey(), nil).AnyTimes()
					// empty vote extension
					signedExtension := validator.SignVoteExtension(t, []byte{}, ec.Ctx.HeaderInfo().Height-1, ec.Ctx.ChainID())
					signedVoteExtensions = append(signedVoteExtensions, signedExtension)
				}

				return &Scenario{
					TotalPower:   totalPower,
					ValidatorSet: validatorAndExtensions.Vals,
					Extensions:   signedVoteExtensions,
				}
			},
			expectError: true,
		},
		{
			name: "1/3 of validators provided invalid bls signature",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				// each validator has the same voting power
				numValidators := 9
				invalidValidBlsSig := numValidators / 3

				validatorAndExtensions, totalPower := generateNValidatorAndVoteExtensions(t, numValidators, &bh, ec.Epoch.EpochNumber)

				var signedVoteExtensions []cbftt.ExtendedVoteInfo
				for i, val := range validatorAndExtensions.Vals {
					validator := val
					ek.EXPECT().GetPubKeyByConsAddr(gomock.Any(), sdk.ConsAddress(validator.ValidatorAddress(t).Bytes())).Return(validator.ProtoPubkey(), nil).AnyTimes()

					if i < invalidValidBlsSig {
						ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(checkpointingtypes.ErrInvalidBlsSignature).AnyTimes()
					} else {
						ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(nil).AnyTimes()
					}
					ek.EXPECT().GetBlsPubKey(gomock.Any(), validator.ValidatorAddress(t)).Return(validator.BlsPubKey(), nil).AnyTimes()
					marshaledExtension, err := validatorAndExtensions.Extensions[i].Marshal()
					require.NoError(t, err)
					signedExtension := validator.SignVoteExtension(t, marshaledExtension, ec.Ctx.HeaderInfo().Height-1, ec.Ctx.ChainID())
					signedVoteExtensions = append(signedVoteExtensions, signedExtension)
				}

				return &Scenario{
					TotalPower:   totalPower,
					ValidatorSet: validatorAndExtensions.Vals,
					Extensions:   signedVoteExtensions,
				}
			},
			expectError: true,
		},
		{
			name: "less than 1/3 of validators provided invalid bls signature",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				// each validator has the same voting power
				numValidators := 9
				invalidValidBlsSig := numValidators/3 - 1

				validatorAndExtensions, totalPower := generateNValidatorAndVoteExtensions(t, numValidators, &bh, ec.Epoch.EpochNumber)

				var signedVoteExtensions []cbftt.ExtendedVoteInfo
				for i, val := range validatorAndExtensions.Vals {
					validator := val
					ek.EXPECT().GetPubKeyByConsAddr(gomock.Any(), sdk.ConsAddress(validator.ValidatorAddress(t).Bytes())).Return(validator.ProtoPubkey(), nil).AnyTimes()

					if i < invalidValidBlsSig {
						ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(checkpointingtypes.ErrInvalidBlsSignature).AnyTimes()
					} else {
						ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(nil).AnyTimes()
					}
					ek.EXPECT().GetBlsPubKey(gomock.Any(), validator.ValidatorAddress(t)).Return(validator.BlsPubKey(), nil).AnyTimes()
					marshaledExtension, err := validatorAndExtensions.Extensions[i].Marshal()
					require.NoError(t, err)
					signedExtension := validator.SignVoteExtension(t, marshaledExtension, ec.Ctx.HeaderInfo().Height-1, ec.Ctx.ChainID())
					signedVoteExtensions = append(signedVoteExtensions, signedExtension)
				}

				return &Scenario{
					TotalPower:   totalPower,
					ValidatorSet: validatorAndExtensions.Vals,
					Extensions:   signedVoteExtensions,
				}
			},
			expectError: false,
		},
		{
			name: "2/3 + 1 of validators voted for valid block hash, the rest voted for invalid block hash",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				bh1 := randomBlockHash()

				validatorAndExtensionsValid, totalPowerValid := generateNValidatorAndVoteExtensions(t, 7, &bh, ec.Epoch.EpochNumber)
				validatorAndExtensionsInvalid, totalPowerInvalid := generateNValidatorAndVoteExtensions(t, 2, &bh1, ec.Epoch.EpochNumber)

				var allvalidators []TestValidator
				allvalidators = append(allvalidators, validatorAndExtensionsValid.Vals...)
				allvalidators = append(allvalidators, validatorAndExtensionsInvalid.Vals...)

				var allExtensions []checkpointingtypes.VoteExtension
				allExtensions = append(allExtensions, validatorAndExtensionsValid.Extensions...)
				allExtensions = append(allExtensions, validatorAndExtensionsInvalid.Extensions...)

				var signedVoteExtensions []cbftt.ExtendedVoteInfo
				for i, val := range allvalidators {
					validator := val
					ek.EXPECT().GetPubKeyByConsAddr(gomock.Any(), sdk.ConsAddress(validator.ValidatorAddress(t).Bytes())).Return(validator.ProtoPubkey(), nil).AnyTimes()
					ek.EXPECT().VerifyBLSSig(gomock.Any(), allExtensions[i].ToBLSSig()).Return(nil).AnyTimes()
					ek.EXPECT().GetBlsPubKey(gomock.Any(), validator.ValidatorAddress(t)).Return(validator.BlsPubKey(), nil).AnyTimes()
					marshaledExtension, err := allExtensions[i].Marshal()
					require.NoError(t, err)
					signedExtension := validator.SignVoteExtension(t, marshaledExtension, ec.Ctx.HeaderInfo().Height-1, ec.Ctx.ChainID())
					signedVoteExtensions = append(signedVoteExtensions, signedExtension)
				}

				return &Scenario{
					TotalPower:   totalPowerValid + totalPowerInvalid,
					ValidatorSet: allvalidators,
					Extensions:   signedVoteExtensions,
				}
			},
			expectError: false,
		},
		{
			name: "All valid vote extensions",
			scenarioSetup: func(ec *EpochAndCtx, ek *mocks.MockCheckpointingKeeper) *Scenario {
				bh := randomBlockHash()
				validatorAndExtensions, totalPower := generateNValidatorAndVoteExtensions(t, 4, &bh, ec.Epoch.EpochNumber)

				var signedVoteExtensions []cbftt.ExtendedVoteInfo
				for i, val := range validatorAndExtensions.Vals {
					validator := val
					ek.EXPECT().GetPubKeyByConsAddr(gomock.Any(), sdk.ConsAddress(validator.ValidatorAddress(t).Bytes())).Return(validator.ProtoPubkey(), nil).AnyTimes()
					ek.EXPECT().VerifyBLSSig(gomock.Any(), validatorAndExtensions.Extensions[i].ToBLSSig()).Return(nil).AnyTimes()
					ek.EXPECT().GetBlsPubKey(gomock.Any(), validator.ValidatorAddress(t)).Return(validator.BlsPubKey(), nil).AnyTimes()
					marshaledExtension, err := validatorAndExtensions.Extensions[i].Marshal()
					require.NoError(t, err)
					signedExtension := validator.SignVoteExtension(t, marshaledExtension, ec.Ctx.HeaderInfo().Height-1, ec.Ctx.ChainID())
					signedVoteExtensions = append(signedVoteExtensions, signedExtension)
				}

				return &Scenario{
					TotalPower:   totalPower,
					ValidatorSet: validatorAndExtensions.Vals,
					Extensions:   signedVoteExtensions,
				}
			},
			expectError: false,
		},

		// TODO: Add scenarios testing compatibility of prepareProposal, processProposal and preBlocker
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			ek := mocks.NewMockCheckpointingKeeper(c)
			mem := mempool.NoOpMempool{}
			ec := epochAndVoteExtensionCtx()
			scenario := tt.scenarioSetup(ec, ek)
			// Those are true for every scenario
			ek.EXPECT().GetEpoch(gomock.Any()).Return(ec.Epoch).AnyTimes()
			ek.EXPECT().GetTotalVotingPower(gomock.Any(), ec.Epoch.EpochNumber).Return(scenario.TotalPower).AnyTimes()
			ek.EXPECT().GetValidatorSet(gomock.Any(), ec.Epoch.EpochNumber).Return(et.NewSortedValidatorSet(ToValidatorSet(scenario.ValidatorSet))).AnyTimes()

			h := checkpointing.NewProposalHandler(
				log.NewNopLogger(),
				ek,
				mem,
				nil,
			)
			prepareProposalFn := h.PrepareProposal()

			prop, err := prepareProposalFn(ec.Ctx, requestPrepareProposal(ec.Ctx.HeaderInfo().Height, scenario.Extensions))
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, prop.Txs, 1)
				var checkpoint checkpointingtypes.InjectedCheckpoint
				err := checkpoint.Unmarshal(prop.Txs[0])
				require.NoError(t, err)
				err = verifyCheckpoint(scenario.ValidatorSet, checkpoint.Ckpt.Ckpt)
				require.NoError(t, err)
			}
		})
	}
}
