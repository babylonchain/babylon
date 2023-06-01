package configurer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer/chain"
	"github.com/babylonchain/babylon/test/e2e/configurer/config"
	"github.com/babylonchain/babylon/test/e2e/containers"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
	"github.com/stretchr/testify/require"
)

// baseConfigurer is the base implementation for the
// other 2 types of configurers. It is not meant to be used
// on its own. Instead, it is meant to be embedded
// by composition into more concrete configurers.
type baseConfigurer struct {
	chainConfigs     []*chain.Config
	containerManager *containers.Manager
	setupTests       setupFn
	syncUntilHeight  int64 // the height until which to wait for validators to sync when first started.
	t                *testing.T
}

// defaultSyncUntilHeight arbitrary small height to make sure the chain is making progress.
const defaultSyncUntilHeight = 3

func (bc *baseConfigurer) ClearResources() error {
	bc.t.Log("tearing down e2e integration test suite...")

	if err := bc.containerManager.ClearResources(); err != nil {
		bc.t.Errorf("failed to clean resources: %v", err)
		return err
	}

	for _, chainConfig := range bc.chainConfigs {
		if err := os.RemoveAll(chainConfig.DataDir); err != nil {
			bc.t.Errorf("failed to remove folder %s for chain %s: %v", chainConfig.DataDir, chainConfig.Id, err)
			return err
		}
	}

	return nil
}

func (bc *baseConfigurer) GetChainConfig(chainIndex int) *chain.Config {
	return bc.chainConfigs[chainIndex]
}

func (bc *baseConfigurer) RunValidators() error {
	for _, chainConfig := range bc.chainConfigs {
		if err := bc.runValidators(chainConfig); err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConfigurer) runValidators(chainConfig *chain.Config) error {
	bc.t.Logf("starting %s validator containers...", chainConfig.Id)
	for _, node := range chainConfig.NodeConfigs {
		if err := node.Run(); err != nil {
			return err
		}
	}
	return nil
}

// RunIBC runs a relayer between every possible pair of chains.
func (bc *baseConfigurer) RunIBC() error {
	for i := 0; i < len(bc.chainConfigs); i++ {
		for j := i + 1; j < len(bc.chainConfigs); j++ {
			chainConfigA := bc.chainConfigs[i]
			chainConfigB := bc.chainConfigs[j]
			channelCfg := config.NewIBCChannelConfigTwoBabylonChains(chainConfigA.Id, chainConfigB.Id)
			if err := bc.runIBCRelayer(chainConfigA, chainConfigB, channelCfg); err != nil {
				return err
			}
		}
	}
	return nil
}

// runIBCRelayer runs a relayer between a given pair of chains with a given
// config for the IBC channel
func (bc *baseConfigurer) runIBCRelayer(chainConfigA *chain.Config, chainConfigB *chain.Config, channelCfg *config.IBCChannelConfig) error {
	bc.t.Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "bbn-e2e-testnet-hermes-")
	if err != nil {
		return err
	}

	hermesCfgPath := path.Join(tmpDir, "hermes")

	if err := os.MkdirAll(hermesCfgPath, 0o755); err != nil {
		return err
	}

	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	if err != nil {
		return err
	}

	// we are using non validator nodes as validator are constantly sending bls
	// transactions, which makes relayer operations failing
	relayerNodeA := chainConfigA.NodeConfigs[2]
	relayerNodeB := chainConfigB.NodeConfigs[2]

	hermesResource, err := bc.containerManager.RunHermesResource(
		chainConfigA.Id,
		relayerNodeA.Name,
		relayerNodeA.Mnemonic,
		chainConfigB.Id,
		relayerNodeB.Name,
		relayerNodeB.Mnemonic,
		hermesCfgPath)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/state", hermesResource.GetHostPort("3031/tcp"))

	require.Eventually(bc.t, func() bool {
		resp, err := http.Get(endpoint)
		if err != nil {
			return false
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}

		var respBody map[string]interface{}
		if err := json.Unmarshal(bz, &respBody); err != nil {
			return false
		}

		status, ok := respBody["status"].(string)
		require.True(bc.t, ok)
		result, ok := respBody["result"].(map[string]interface{})
		require.True(bc.t, ok)

		chains, ok := result["chains"].([]interface{})
		require.True(bc.t, ok)

		return status == "success" && len(chains) == 2
	},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy")

	bc.t.Logf("started Hermes relayer container: %s", hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two babylon chains
	return bc.ConnectIBCChains(channelCfg)
}

func (bc *baseConfigurer) ConnectIBCChains(cfg *config.IBCChannelConfig) error {
	bc.t.Logf("connecting %s and %s chains via IBC", cfg.ChainAID, cfg.ChainBID)
	cmd := cfg.ToCmd()
	_, _, err := bc.containerManager.ExecHermesCmd(bc.t, cmd, "SUCCESS")
	if err != nil {
		return err
	}
	bc.t.Logf("connected %s and %s chains via IBC", cfg.ChainAID, cfg.ChainBID)
	return nil
}

// DeployWasmContract instantiates a wasm smart contract from a given
// contract code and a given instantiation message on a given chain
// It returns the contract address
func (bc *baseConfigurer) DeployWasmContract(contractCodePath string, chain *chain.Config, initMsg string) (string, error) {
	nonValidatorNode, err := chain.GetNodeAtIndex(2) // TODO: find a generic way to get non validator node
	if err != nil {
		return "", err
	}

	// store wasm code
	nonValidatorNode.StoreWasmCode(contractCodePath, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlocks(3)

	// instantiate contract with the wasm code ID
	latestWasmCodeID := int(nonValidatorNode.QueryLatestWasmCodeID())
	if latestWasmCodeID == 0 {
		return "", fmt.Errorf("StoreWasmCode failed")
	}

	nonValidatorNode.InstantiateWasmContract(
		strconv.Itoa(latestWasmCodeID),
		initMsg,
		initialization.ValidatorWalletName,
	)
	nonValidatorNode.WaitForNextBlocks(3)

	// get the address of the instantiated contract
	contracts, err := nonValidatorNode.QueryContractsFromId(latestWasmCodeID)
	if err != nil {
		return "", err
	}
	if len(contracts) == 0 {
		return "", fmt.Errorf("InstantiateWasmContract failed")
	}
	contractAddr := contracts[0]

	return contractAddr, nil
}

func (bc *baseConfigurer) initializeChainConfigFromInitChain(initializedChain *initialization.Chain, chainConfig *chain.Config) {
	chainConfig.ChainMeta = initializedChain.ChainMeta
	chainConfig.NodeConfigs = make([]*chain.NodeConfig, 0, len(initializedChain.Nodes))
	setupTime := time.Now()
	for i, validator := range initializedChain.Nodes {
		conf := chain.NewNodeConfig(bc.t, validator, chainConfig.ValidatorInitConfigs[i], chainConfig.Id, bc.containerManager).WithSetupTime(setupTime)
		chainConfig.NodeConfigs = append(chainConfig.NodeConfigs, conf)
	}
}
