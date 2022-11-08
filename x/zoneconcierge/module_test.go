package zoneconcierge_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/app"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clientexported "github.com/cosmos/ibc-go/v5/modules/core/02-client/exported"
	"github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v5/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	ibctestingmock "github.com/cosmos/ibc-go/v5/testing/mock"
	"github.com/cosmos/ibc-go/v5/testing/simapp"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// ZoneConciergeTestSuite provides a test suite for IBC functionalities in ZoneConcierge
// (adapted from https://github.com/cosmos/ibc-go/blob/v5.0.0/modules/core/02-client/keeper/keeper_test.go)
type ZoneConciergeTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	babylonChain *ibctesting.TestChain
	czChain      *ibctesting.TestChain

	// System states of the simulated Babylon chain
	cdc            codec.Codec
	ctx            sdk.Context
	keeper         clientexported.ClientKeeper
	zcKeeper       zckeeper.Keeper
	consensusState *ibctmtypes.ConsensusState
	header         *ibctmtypes.Header
	valSet         *tmtypes.ValidatorSet
	valSetHash     tmbytes.HexBytes
	privVal        tmtypes.PrivValidator
	now            time.Time
	past           time.Time
	signers        map[string]tmtypes.PrivValidator
}

func TestZoneConciergeTestSuite(t *testing.T) {
	suite.Run(t, new(ZoneConciergeTestSuite))
}

func (suite *ZoneConciergeTestSuite) SetupTest() {
	// set up 2 Test chains with default constructors
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	// replace the first test chain with a Babylon chain
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		babylonApp := app.Setup(suite.T(), false)
		suite.zcKeeper = babylonApp.ZoneConciergeKeeper
		encCdc := app.MakeTestEncodingConfig()
		return babylonApp, app.NewDefaultGenesisState(encCdc.Marshaler)
	}
	babylonChainID := ibctesting.GetChainID(1)
	suite.coordinator.Chains[babylonChainID] = ibctesting.NewTestChain(suite.T(), suite.coordinator, babylonChainID)

	suite.babylonChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.czChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.past = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	now2 := suite.now.Add(time.Hour)
	app := simapp.Setup(false)

	babylonChainHeight := uint64(5) // TODO: find out why it's 5 (any value > 0 is okay)
	suite.cdc = app.AppCodec()
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: int64(babylonChainHeight), ChainID: babylonChainID, Time: now2})
	suite.keeper = app.IBCKeeper.ClientKeeper
	suite.privVal = ibctestingmock.NewPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	testClientHeightMinus1 := types.NewHeight(0, babylonChainHeight-1)

	validator := tmtypes.NewValidator(pubKey, 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	suite.valSetHash = suite.valSet.Hash()
	suite.signers = make(map[string]tmtypes.PrivValidator, 1)
	suite.signers[validator.Address.String()] = suite.privVal

	testClientHeight := clienttypes.NewHeight(0, babylonChainHeight)
	testChainID := ibctesting.GetChainID(2)
	suite.header = suite.babylonChain.CreateTMClientHeader(testChainID, int64(testClientHeight.RevisionHeight), testClientHeightMinus1, now2, suite.valSet, suite.valSet, suite.valSet, suite.signers)
	suite.consensusState = ibctmtypes.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot([]byte("hash")), suite.valSetHash)

	var validators stakingtypes.Validators
	for i := 1; i < 11; i++ {
		privVal := ibctestingmock.NewPV()
		tmPk, err := privVal.GetPubKey()
		suite.Require().NoError(err)
		pk, err := cryptocodec.FromTmPubKeyInterface(tmPk)
		suite.Require().NoError(err)
		val, err := stakingtypes.NewValidator(sdk.ValAddress(pk.Address()), pk, stakingtypes.Description{})
		suite.Require().NoError(err)

		val.Status = stakingtypes.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)

		hi := stakingtypes.NewHistoricalInfo(suite.ctx.BlockHeader(), validators, sdk.DefaultPowerReduction)
		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), &hi)
	}
}

// TestUpdateClientTendermint provides tests on verifying different headers from the IBC light client
// (https://github.com/cosmos/ibc-go/blob/v5.0.0/modules/core/02-client/keeper/client_test.go)
func (suite *ZoneConciergeTestSuite) TestUpdateClientTendermint() {
	var (
		path         *ibctesting.Path
		updateHeader *ibctmtypes.Header
	)

	// Must create header creation functions since suite.header gets recreated on each test case
	createFutureUpdateFn := func(trustedHeight clienttypes.Height) *ibctmtypes.Header {
		header, err := suite.babylonChain.ConstructUpdateTMClientHeaderWithTrustedHeight(path.EndpointB.Chain, path.EndpointA.ClientID, trustedHeight)
		suite.Require().NoError(err)
		return header
	}
	createPastUpdateFn := func(fillHeight, trustedHeight clienttypes.Height) *ibctmtypes.Header {
		consState, found := suite.babylonChain.App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(suite.babylonChain.GetContext(), path.EndpointA.ClientID, trustedHeight)
		suite.Require().True(found)

		return suite.czChain.CreateTMClientHeader(suite.czChain.ChainID, int64(fillHeight.RevisionHeight), trustedHeight, consState.(*ibctmtypes.ConsensusState).Timestamp.Add(time.Second*5),
			suite.czChain.Vals, suite.czChain.Vals, suite.czChain.Vals, suite.czChain.Signers)
	}

	cases := []struct {
		name                 string
		malleate             func()
		expPass              bool
		expDishonestMajority bool
	}{
		{"valid update", func() {
			clientState := path.EndpointA.GetClientState().(*ibctmtypes.ClientState)
			trustHeight := clientState.GetLatestHeight().(clienttypes.Height)

			// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
			err := path.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			updateHeader = createFutureUpdateFn(trustHeight)
		}, true, false},
		{"valid past update", func() {
			clientState := path.EndpointA.GetClientState()
			trustedHeight := clientState.GetLatestHeight().(clienttypes.Height)

			currHeight := suite.czChain.CurrentHeader.Height
			fillHeight := clienttypes.NewHeight(clientState.GetLatestHeight().GetRevisionNumber(), uint64(currHeight))

			// commit a couple blocks to allow client to fill in gaps
			suite.coordinator.CommitBlock(suite.czChain) // this height is not filled in yet
			suite.coordinator.CommitBlock(suite.czChain) // this height is filled in by the update below

			err := path.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			// ensure fill height not set
			_, found := suite.babylonChain.App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(suite.babylonChain.GetContext(), path.EndpointA.ClientID, fillHeight)
			suite.Require().False(found)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(fillHeight, trustedHeight)
		}, true, false},
		{"valid duplicate update", func() {
			clientID := path.EndpointA.ClientID

			height1 := clienttypes.NewHeight(0, 1)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past,
				NextValidatorsHash: suite.czChain.Vals.Hash(),
			}
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, height1, prevConsState)

			height5 := clienttypes.NewHeight(0, 5)
			// store next consensus state to check that trustedHeight does not need to be hightest consensus state before header height
			nextConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past.Add(time.Minute),
				NextValidatorsHash: suite.czChain.Vals.Hash(),
			}
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, height5, nextConsState)

			height3 := clienttypes.NewHeight(0, 3)
			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(height3, height1)
			// set updateHeader's consensus state in store to create duplicate UpdateClient scenario
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, updateHeader.GetHeight(), updateHeader.ConsensusState())
		}, true, false},
		{"misbehaviour detection: conflicting header", func() {
			clientID := path.EndpointA.ClientID

			height1 := clienttypes.NewHeight(0, 1)
			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past,
				NextValidatorsHash: suite.czChain.Vals.Hash(),
			}
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, height1, prevConsState)

			height5 := clienttypes.NewHeight(0, 5)
			// store next consensus state to check that trustedHeight does not need to be hightest consensus state before header height
			nextConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past.Add(time.Minute),
				NextValidatorsHash: suite.czChain.Vals.Hash(),
			}
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, height5, nextConsState)

			height3 := clienttypes.NewHeight(0, 3)
			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(height3, height1)
			// set conflicting consensus state in store to create misbehaviour scenario
			conflictConsState := updateHeader.ConsensusState()
			conflictConsState.Root = commitmenttypes.NewMerkleRoot([]byte("conflicting apphash"))
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, updateHeader.GetHeight(), conflictConsState)
		}, false, true}, // Babylon modification: fork headers are rejected before being passed to ClientState, and are recorded in the fork index
		{"misbehaviour detection: monotonic time violation", func() {
			clientState := path.EndpointA.GetClientState().(*ibctmtypes.ClientState)
			clientID := path.EndpointA.ClientID
			trustedHeight := clientState.GetLatestHeight().(clienttypes.Height)

			// store intermediate consensus state at a time greater than updateHeader time
			// this will break time monotonicity
			incrementedClientHeight := clientState.GetLatestHeight().Increment().(clienttypes.Height)
			intermediateConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.coordinator.CurrentTime.Add(2 * time.Hour),
				NextValidatorsHash: suite.czChain.Vals.Hash(),
			}
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.babylonChain.GetContext(), clientID, incrementedClientHeight, intermediateConsState)
			// set iteration key
			clientStore := suite.keeper.ClientStore(suite.ctx, clientID)
			ibctmtypes.SetIterationKey(clientStore, incrementedClientHeight)

			clientState.LatestHeight = incrementedClientHeight
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.babylonChain.GetContext(), clientID, clientState)

			updateHeader = createFutureUpdateFn(trustedHeight)
		}, false, true}, // Babylon modification: non-monotonic headers are rejected before being passed to ClientState, and are recorded in the fork index
		{"client state not found", func() {
			updateHeader = createFutureUpdateFn(path.EndpointA.GetClientState().GetLatestHeight().(clienttypes.Height))

			path.EndpointA.ClientID = ibctesting.InvalidID
		}, false, false},
		{"consensus state not found", func() {
			clientState := path.EndpointA.GetClientState()
			tmClient, ok := clientState.(*ibctmtypes.ClientState)
			suite.Require().True(ok)
			tmClient.LatestHeight = tmClient.LatestHeight.Increment().(clienttypes.Height)

			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.babylonChain.GetContext(), path.EndpointA.ClientID, clientState)
			updateHeader = createFutureUpdateFn(clientState.GetLatestHeight().(clienttypes.Height))
		}, false, false},
		{"client is not active", func() {
			clientState := path.EndpointA.GetClientState().(*ibctmtypes.ClientState)
			clientState.FrozenHeight = clienttypes.NewHeight(0, 1)
			suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.babylonChain.GetContext(), path.EndpointA.ClientID, clientState)
			updateHeader = createFutureUpdateFn(clientState.GetLatestHeight().(clienttypes.Height))
		}, false, false},
		{"invalid header", func() {
			updateHeader = createFutureUpdateFn(path.EndpointA.GetClientState().GetLatestHeight().(clienttypes.Height))
			updateHeader.TrustedHeight = updateHeader.TrustedHeight.Increment().(clienttypes.Height)
		}, false, false},
	}

	for _, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path = ibctesting.NewPath(suite.babylonChain, suite.czChain)
			suite.coordinator.SetupClients(path)

			tc.malleate()

			var clientState exported.ClientState
			if tc.expPass {
				clientState = path.EndpointA.GetClientState()
			}

			err := suite.babylonChain.App.GetIBCKeeper().ClientKeeper.UpdateClient(suite.babylonChain.GetContext(), path.EndpointA.ClientID, updateHeader)

			if tc.expPass {
				suite.Require().NoError(err, err)

				newClientState := path.EndpointA.GetClientState()

				// Babylon check: Babylon's IBC light client should never freeze
				suite.Require().True(newClientState.(*ibctmtypes.ClientState).FrozenHeight.IsZero(), "Babylon's IBC light client should never freeze")

				expConsensusState := &ibctmtypes.ConsensusState{
					Timestamp:          updateHeader.GetTime(),
					Root:               commitmenttypes.NewMerkleRoot(updateHeader.Header.GetAppHash()),
					NextValidatorsHash: updateHeader.Header.NextValidatorsHash,
				}

				consensusState, found := suite.babylonChain.App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(suite.babylonChain.GetContext(), path.EndpointA.ClientID, updateHeader.GetHeight())
				suite.Require().True(found)

				// Determine if clientState should be updated or not
				if updateHeader.GetHeight().GT(clientState.GetLatestHeight()) {
					// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
					suite.Require().Equal(updateHeader.GetHeight(), newClientState.GetLatestHeight(), "clientstate height did not update")
				} else {
					// Update will add past consensus state, clientState should not be updated at all
					suite.Require().Equal(clientState.GetLatestHeight(), newClientState.GetLatestHeight(), "client state height updated for past header")
				}

				suite.Require().NoError(err)
				suite.Require().Equal(expConsensusState, consensusState, "consensus state should have been updated on case %s", tc.name)

				/* Extra Babylon checks */
				ctx := suite.babylonChain.GetContext()
				czChainID := suite.czChain.ChainID
				updateHeaderHeight := uint64(updateHeader.Header.Height)
				// updateHeader should be correctly recorded in canonical chain indexer
				expUpdateHeader, err := suite.zcKeeper.GetHeader(ctx, czChainID, updateHeaderHeight)
				suite.Require().NoError(err)
				suite.Require().Equal(expUpdateHeader.Hash, updateHeader.Header.LastCommitHash)
				suite.Require().Equal(expUpdateHeader.Height, updateHeaderHeight)
				// updateHeader should be correctly recorded in chain info indexer
				chainInfo := suite.zcKeeper.GetChainInfo(ctx, czChainID)
				suite.Require().Equal(chainInfo.LatestHeader.Hash, updateHeader.Header.LastCommitHash)
				suite.Require().Equal(chainInfo.LatestHeader.Height, updateHeaderHeight)

			} else {
				if tc.expDishonestMajority {
					/* Extra Babylon checks */
					ctx := suite.babylonChain.GetContext()
					czChainID := suite.czChain.ChainID
					updateHeaderHeight := uint64(updateHeader.Header.Height)
					// updateHeader should be correctly recorded in fork indexer
					expForks := suite.zcKeeper.GetForks(ctx, czChainID, updateHeaderHeight)
					suite.Require().Equal(1, len(expForks.Headers))
					suite.Require().Equal(expForks.Headers[0].Hash, updateHeader.Header.LastCommitHash)
					suite.Require().Equal(expForks.Headers[0].Height, updateHeaderHeight)
					// updateHeader should be correctly recorded in chain info indexer
					chainInfo := suite.zcKeeper.GetChainInfo(ctx, czChainID)
					suite.Require().Equal(1, len(chainInfo.LatestForks.Headers))
					suite.Require().Equal(chainInfo.LatestForks.Headers[0].Hash, updateHeader.Header.LastCommitHash)
					suite.Require().Equal(chainInfo.LatestForks.Headers[0].Height, updateHeaderHeight)
				} else {
					suite.Require().Error(err)
				}
			}
		})
	}
}
