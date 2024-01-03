package configurer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/babylonchain/babylon/types"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/test/e2e/configurer/chain"
	"github.com/babylonchain/babylon/test/e2e/containers"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
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
		return err
	}

	for _, chainConfig := range bc.chainConfigs {
		if err := os.RemoveAll(chainConfig.DataDir); err != nil {
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

func (bc *baseConfigurer) InstantiateBabylonContract() error {
	// Store the contract on the second chain (B)
	chainConfig := bc.chainConfigs[1]
	contractPath := "/bytecode/babylon_contract.wasm"
	nonValidatorNode, err := chainConfig.GetNodeAtIndex(2)
	if err != nil {
		bc.t.Logf("error getting non-validator node: %v", err)
		return err
	}
	nonValidatorNode.StoreWasmCode(contractPath, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()
	latestWasmId := int(nonValidatorNode.QueryLatestWasmCodeID())

	// Instantiate the contract
	// TODO: Get this from the chain config
	babylonTag := "[1,2,3,4]"
	initMsg := fmt.Sprintf(`{ "network": %q, "babylon_tag": %q, "btc_confirmation_depth": %d, "checkpoint_finalization_timeout": %d, "notify_cosmos_zone": %s }`,
		types.BtcRegtest,
		base64.StdEncoding.EncodeToString([]byte(babylonTag)),
		1,
		2,
		"false",
	)
	nonValidatorNode.InstantiateWasmContract(
		strconv.Itoa(latestWasmId),
		initMsg,
		initialization.ValidatorWalletName,
	)
	nonValidatorNode.WaitForNextBlock()
	contracts, err := nonValidatorNode.QueryContractsFromId(1)
	if err != nil {
		bc.t.Logf("error querying contracts from id: %v", err)
		return err
	}
	require.Len(bc.t, contracts, 1, "Wrong number of contracts for the counter")
	contractAddr := contracts[0]

	// Set the contract address in the IBC chain config port id.
	chainConfig.IBCConfig.PortID = fmt.Sprintf("wasm.%s", contractAddr)

	return nil
}

func (bc *baseConfigurer) RunIBC() error {
	// Run a relayer between every possible pair of chains.
	for i := 0; i < len(bc.chainConfigs); i++ {
		for j := i + 1; j < len(bc.chainConfigs); j++ {
			if err := bc.runIBCRelayer(bc.chainConfigs[i], bc.chainConfigs[j]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *baseConfigurer) runIBCRelayer(chainConfigA *chain.Config, chainConfigB *chain.Config) error {
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
	time.Sleep(3 * time.Second)

	// create the client, connection and channel between the two babylon chains
	return bc.connectIBCChains(chainConfigA, chainConfigB)
}

func (bc *baseConfigurer) connectIBCChains(chainA *chain.Config, chainB *chain.Config) error {
	bc.t.Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	require.Equal(bc.t, chainA.IBCConfig.Order, chainB.IBCConfig.Order)
	require.Equal(bc.t, chainA.IBCConfig.Version, chainB.IBCConfig.Version)
	cmd := []string{"hermes", "create", "channel",
		"--a-chain", chainA.ChainMeta.Id, "--b-chain", chainB.ChainMeta.Id, // channel ID
		"--a-port", chainA.IBCConfig.PortID, "--b-port", chainB.IBCConfig.PortID, // port
		"--order", chainA.IBCConfig.Order.String(),
		"--channel-version", chainA.IBCConfig.Version,
		"--new-client-connection", "--yes",
	}
	_, _, err := bc.containerManager.ExecHermesCmd(bc.t, cmd, "SUCCESS")
	if err != nil {
		return err
	}
	bc.t.Logf("connected %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	return nil
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
