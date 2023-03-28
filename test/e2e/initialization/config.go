package initialization

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/babylonchain/babylon/privval"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	blctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/babylonchain/babylon/test/e2e/util"
)

// NodeConfig is a confiuration for the node supplied from the test runner
// to initialization scripts. It should be backwards compatible with earlier
// versions. If this struct is updated, the change must be backported to earlier
// branches that might be used for upgrade testing.
type NodeConfig struct {
	Name               string // name of the config that will also be assigned to Docke container.
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
	IsValidator        bool   // flag indicating whether a node should be a validator
}

const (
	// common
	BabylonDenom                 = "ubbn"
	MinGasPrice                  = "0.000"
	ValidatorWalletName          = "val"
	BabylonOpReturnTag           = "bbni"
	BabylonBtcConfirmationPeriod = 2
	BabylonBtcFinalizationPeriod = 4
	// chainA
	ChainAID        = "bbn-test-a"
	BabylonBalanceA = 200000000000
	StakeAmountA    = 100000000000
	// chainB
	ChainBID        = "bbn-test-b"
	BabylonBalanceB = 500000000000
	StakeAmountB    = 400000000000

	EpochDuration         = time.Second * 60
	TWAPPruningKeepPeriod = EpochDuration / 4
)

var (
	StakeAmountIntA  = sdk.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(BabylonDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(BabylonDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%d%s", BabylonBalanceA, BabylonDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s", BabylonBalanceB, BabylonDenom)
)

func addAccount(path, moniker, amountStr string, accAddr sdk.AccAddress, forkHeight int) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
	genAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

	// TODO: Make the SDK make it far cleaner to add an account to GenesisState
	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	genDoc.InitialHeight = int64(forkHeight)

	authGenState := authtypes.GetGenesisStateFromAppState(util.Cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	if accs.Contains(accAddr) {
		return fmt.Errorf("failed to add account to genesis state; account already exists: %s", accAddr)
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	accs = append(accs, genAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := util.Cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(util.Cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := util.Cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}

	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

//nolint:typecheck
func updateModuleGenesis[V proto.Message](appGenState map[string]json.RawMessage, moduleName string, protoVal V, updateGenesis func(V)) error {
	if err := util.Cdc.UnmarshalJSON(appGenState[moduleName], protoVal); err != nil {
		return err
	}
	updateGenesis(protoVal)
	newGenState := protoVal

	bz, err := util.Cdc.MarshalJSON(newGenState)
	if err != nil {
		return err
	}
	appGenState[moduleName] = bz
	return nil
}

func initGenesis(chain *internalChain, votingPeriod, expeditedVotingPeriod time.Duration, forkHeight int) error {
	// initialize a genesis file
	configDir := chain.nodes[0].configDir()

	for _, val := range chain.nodes {
		addr, err := val.keyInfo.GetAddress()

		if err != nil {
			return err
		}

		if chain.chainMeta.Id == ChainAID {
			if err := addAccount(configDir, "", InitBalanceStrA, addr, forkHeight); err != nil {
				return err
			}
		} else if chain.chainMeta.Id == ChainBID {
			if err := addAccount(configDir, "", InitBalanceStrB, addr, forkHeight); err != nil {
				return err
			}
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range chain.nodes[1:] {
		_, err := util.CopyFile(
			filepath.Join(configDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		if err != nil {
			return err
		}
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(chain.nodes[0].configDir())
	config.Moniker = chain.nodes[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, banktypes.ModuleName, &banktypes.GenesisState{}, updateBankGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, staketypes.ModuleName, &staketypes.GenesisState{}, updateStakeGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, crisistypes.ModuleName, &crisistypes.GenesisState{}, updateCrisisGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, genutiltypes.ModuleName, &genutiltypes.GenesisState{}, updateGenUtilGenesis(chain))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, checkpointingtypes.ModuleName, checkpointingtypes.DefaultGenesis(), updateCheckpointingGenesis(chain))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, blctypes.ModuleName, blctypes.DefaultGenesis(), updateBtcLightClientGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, btccheckpointtypes.ModuleName, btccheckpointtypes.DefaultGenesis(), updateBtccheckpointGenesis)
	if err != nil {
		return err
	}

	bz, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc.AppState = bz

	genesisJson, err := tmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	// write the updated genesis file to each validator
	for _, val := range chain.nodes {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), genesisJson); err != nil {
			return err
		}
	}
	return nil
}

func updateBankGenesis(bankGenState *banktypes.GenesisState) {
	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     BabylonDenom,
		Base:        BabylonDenom,
		Symbol:      BabylonDenom,
		Name:        BabylonDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BabylonDenom,
				Exponent: 0,
			},
		},
	})
}

func updateStakeGenesis(stakeGenState *staketypes.GenesisState) {
	stakeGenState.Params = staketypes.Params{
		BondDenom:         BabylonDenom,
		MaxValidators:     100,
		MaxEntries:        7,
		HistoricalEntries: 10000,
		UnbondingTime:     240000000000,
		MinCommissionRate: sdk.ZeroDec(),
	}
}

func updateCrisisGenesis(crisisGenState *crisistypes.GenesisState) {
	crisisGenState.ConstantFee.Denom = BabylonDenom
}

func updateBtcLightClientGenesis(blcGenState *blctypes.GenesisState) {
	btcSimnetGenesisHex := "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000"
	baseBtcHeader, err := bbn.NewBTCHeaderBytesFromHex(btcSimnetGenesisHex)
	if err != nil {
		panic(err)
	}
	work := blctypes.CalcWork(&baseBtcHeader)
	blcGenState.BaseBtcHeader = *blctypes.NewBTCHeaderInfo(&baseBtcHeader, baseBtcHeader.Hash(), 0, &work)
}

func updateBtccheckpointGenesis(btccheckpointGenState *btccheckpointtypes.GenesisState) {
	btccheckpointGenState.Params = btccheckpointtypes.DefaultParams()
	btccheckpointGenState.Params.BtcConfirmationDepth = BabylonBtcConfirmationPeriod
	btccheckpointGenState.Params.CheckpointFinalizationTimeout = BabylonBtcFinalizationPeriod
}

func updateGenUtilGenesis(c *internalChain) func(*genutiltypes.GenesisState) {
	return func(genUtilGenState *genutiltypes.GenesisState) {
		// generate genesis txs
		genTxs := make([]json.RawMessage, 0, len(c.nodes))
		for _, node := range c.nodes {
			if !node.isValidator {
				continue
			}

			stakeAmountCoin := StakeAmountCoinA
			if c.chainMeta.Id != ChainAID {
				stakeAmountCoin = StakeAmountCoinB
			}
			createValmsg, err := node.buildCreateValidatorMsg(stakeAmountCoin)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			signedTx, err := node.signMsg(createValmsg)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			txRaw, err := util.Cdc.MarshalJSON(signedTx)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}
			genTxs = append(genTxs, txRaw)
		}
		genUtilGenState.GenTxs = genTxs
	}
}

func updateCheckpointingGenesis(c *internalChain) func(*checkpointingtypes.GenesisState) {
	return func(checkpointingGenState *checkpointingtypes.GenesisState) {
		var genKeys []*checkpointingtypes.GenesisKey

		for _, node := range c.nodes {
			if !node.isValidator {
				continue
			}

			proofOfPossession, err := privval.BuildPoP(node.consensusKey.PrivKey, node.consensusKey.BlsPrivKey)

			if err != nil {
				panic("It should be possible to build proof of possesion from validator private keys")
			}

			valPubKey, err := cryptocodec.FromTmPubKeyInterface(node.consensusKey.PubKey)

			if err != nil {
				panic("It should be possible to retrieve validator public key")
			}

			da, err := sdk.AccAddressFromBech32(node.consensusKey.DelegatorAddress)

			if err != nil {
				panic("It should be possible to get validator address from delegator address")
			}

			va := sdk.ValAddress(da)

			genKey := &checkpointingtypes.GenesisKey{
				ValidatorAddress: va.String(),
				BlsKey: &checkpointingtypes.BlsKey{
					Pubkey: &node.consensusKey.BlsPubKey,
					Pop:    proofOfPossession,
				},
				ValPubkey: valPubKey.(*ed25519.PubKey),
			}

			genKeys = append(genKeys, genKey)
		}

		checkpointingGenState.GenesisKeys = genKeys
	}
}
