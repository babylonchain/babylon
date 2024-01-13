package chain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/babylonchain/babylon/test/e2e/util"
	blc "github.com/babylonchain/babylon/x/btclightclient/types"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	mtypes "github.com/babylonchain/babylon/x/monitor/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmabcitypes "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) QueryGRPCGateway(path string, queryParams url.Values) ([]byte, error) {
	// add the URL for the given validator ID, and pre-pend to path.
	hostPort, err := n.containerManager.GetHostPort(n.Name, "1317/tcp")
	require.NoError(n.t, err)
	endpoint := fmt.Sprintf("http://%s", hostPort)
	fullQueryPath := fmt.Sprintf("%s/%s", endpoint, path)

	var resp *http.Response
	require.Eventually(n.t, func() bool {
		req, err := http.NewRequest("GET", fullQueryPath, nil)
		if err != nil {
			return false
		}

		if len(queryParams) > 0 {
			req.URL.RawQuery = queryParams.Encode()
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			n.t.Logf("error while executing HTTP request: %s", err.Error())
			return false
		}

		return resp.StatusCode != http.StatusServiceUnavailable
	}, time.Minute, time.Millisecond*10, "failed to execute HTTP request")

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bz))
	}
	return bz, nil
}

// QueryModuleAddress returns the address of a given module
func (n *NodeConfig) QueryModuleAddress(name string) (sdk.AccAddress, error) {
	path := fmt.Sprintf("/cosmos/auth/v1beta1/module_accounts/%s", name)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp authtypes.QueryModuleAccountByNameResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return sdk.AccAddress{}, err
	}
	// cast to account
	var account sdk.AccountI
	if err := util.EncodingConfig.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return sdk.AccAddress{}, err
	}

	return account.GetAddress(), nil
}

// QueryBalances returns balances at the address.
func (n *NodeConfig) QueryBalances(address string) (sdk.Coins, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s", address)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.Coins{}, err
	}
	return balancesResp.GetBalances(), nil
}

func (n *NodeConfig) QuerySupplyOf(denom string) (sdkmath.Int, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/supply/%s", denom)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var supplyResp banktypes.QuerySupplyOfResponse
	if err := util.Cdc.UnmarshalJSON(bz, &supplyResp); err != nil {
		return sdkmath.NewInt(0), err
	}
	return supplyResp.Amount.Amount, nil
}

// QueryBlock gets block at a specific height
func (n *NodeConfig) QueryBlock(height int64) (*tmtypes.Block, error) {
	block, err := n.rpcClient.Block(context.Background(), &height)
	if err != nil {
		return nil, err
	}
	return block.Block, nil
}

// QueryHashFromBlock gets block hash at a specific height. Otherwise, error.
func (n *NodeConfig) QueryHashFromBlock(height int64) (string, error) {
	block, err := n.rpcClient.Block(context.Background(), &height)
	if err != nil {
		return "", err
	}
	return block.BlockID.Hash.String(), nil
}

// QueryCurrentHeight returns the current block height of the node or error.
func (n *NodeConfig) QueryCurrentHeight() (int64, error) {
	status, err := n.rpcClient.Status(context.Background())
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// QueryLatestBlockTime returns the latest block time.
func (n *NodeConfig) QueryLatestBlockTime() time.Time {
	status, err := n.rpcClient.Status(context.Background())
	require.NoError(n.t, err)
	return status.SyncInfo.LatestBlockTime
}

// QueryListSnapshots gets all snapshots currently created for a node.
func (n *NodeConfig) QueryListSnapshots() ([]*tmabcitypes.Snapshot, error) {
	abciResponse, err := n.rpcClient.ABCIQuery(context.Background(), "/app/snapshots", nil)
	if err != nil {
		return nil, err
	}

	var listSnapshots tmabcitypes.ResponseListSnapshots
	if err := json.Unmarshal(abciResponse.Response.Value, &listSnapshots); err != nil {
		return nil, err
	}

	return listSnapshots.Snapshots, nil
}

// func (n *NodeConfig) QueryContractsFromId(codeId int) ([]string, error) {
// 	path := fmt.Sprintf("/cosmwasm/wasm/v1/code/%d/contracts", codeId)
// 	bz, err := n.QueryGRPCGateway(path)

// 	require.NoError(n.t, err)

// 	var contractsResponse wasmtypes.QueryContractsByCodeResponse
// 	if err := util.Cdc.UnmarshalJSON(bz, &contractsResponse); err != nil {
// 		return nil, err
// 	}

// 	return contractsResponse.Contracts, nil
// }

func (n *NodeConfig) QueryRawCheckpoint(epoch uint64) (*ct.RawCheckpointWithMeta, error) {
	path := fmt.Sprintf("babylon/checkpointing/v1/raw_checkpoint/%d", epoch)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var checkpointingResponse ct.QueryRawCheckpointResponse
	if err := util.Cdc.UnmarshalJSON(bz, &checkpointingResponse); err != nil {
		return nil, err
	}

	return checkpointingResponse.RawCheckpoint, nil
}

func (n *NodeConfig) QueryRawCheckpoints(pagination *query.PageRequest) (*ct.QueryRawCheckpointsResponse, error) {
	queryParams := url.Values{}
	if pagination != nil {
		queryParams.Set("pagination.key", base64.URLEncoding.EncodeToString(pagination.Key))
		queryParams.Set("pagination.limit", strconv.Itoa(int(pagination.Limit)))
	}

	bz, err := n.QueryGRPCGateway("babylon/checkpointing/v1/raw_checkpoints", queryParams)
	require.NoError(n.t, err)

	var checkpointingResponse ct.QueryRawCheckpointsResponse
	if err := util.Cdc.UnmarshalJSON(bz, &checkpointingResponse); err != nil {
		return nil, err
	}

	return &checkpointingResponse, nil
}

func (n *NodeConfig) QueryBtcBaseHeader() (*blc.BTCHeaderInfo, error) {
	bz, err := n.QueryGRPCGateway("babylon/btclightclient/v1/baseheader", url.Values{})
	require.NoError(n.t, err)

	var blcResponse blc.QueryBaseHeaderResponse
	if err := util.Cdc.UnmarshalJSON(bz, &blcResponse); err != nil {
		return nil, err
	}

	return blcResponse.Header, nil
}

func (n *NodeConfig) QueryTip() (*blc.BTCHeaderInfo, error) {
	bz, err := n.QueryGRPCGateway("babylon/btclightclient/v1/tip", url.Values{})
	require.NoError(n.t, err)

	var blcResponse blc.QueryTipResponse
	if err := util.Cdc.UnmarshalJSON(bz, &blcResponse); err != nil {
		return nil, err
	}

	return blcResponse.Header, nil
}

func (n *NodeConfig) QueryHeaderDepth(hash string) (uint64, error) {
	path := fmt.Sprintf("babylon/btclightclient/v1/depth/%s", hash)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var blcResponse blc.QueryHeaderDepthResponse
	if err := util.Cdc.UnmarshalJSON(bz, &blcResponse); err != nil {
		return 0, err
	}

	return blcResponse.Depth, nil
}

func (n *NodeConfig) QueryFinalizedChainsInfo(chainIDs []string) ([]*zctypes.FinalizedChainInfo, error) {
	queryParams := url.Values{}
	for _, chainId := range chainIDs {
		queryParams.Add("chain_ids", chainId)
	}

	bz, err := n.QueryGRPCGateway("babylon/zoneconcierge/v1/finalized_chains_info", queryParams)
	require.NoError(n.t, err)

	var resp zctypes.QueryFinalizedChainsInfoResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return resp.FinalizedChainsInfo, nil
}

func (n *NodeConfig) QueryEpochChainsInfo(epochNum uint64, chainIDs []string) ([]*zctypes.ChainInfo, error) {
	queryParams := url.Values{}
	for _, chainId := range chainIDs {
		queryParams.Add("epoch_num", fmt.Sprintf("%d", epochNum))
		queryParams.Add("chain_ids", chainId)
	}

	bz, err := n.QueryGRPCGateway("babylon/zoneconcierge/v1/epoch_chains_info", queryParams)
	require.NoError(n.t, err)

	var resp zctypes.QueryEpochChainsInfoResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return resp.ChainsInfo, nil
}

func (n *NodeConfig) QueryChains() (*[]string, error) {
	bz, err := n.QueryGRPCGateway("babylon/zoneconcierge/v1/chains", url.Values{})
	require.NoError(n.t, err)
	var chainsResponse zctypes.QueryChainListResponse
	if err := util.Cdc.UnmarshalJSON(bz, &chainsResponse); err != nil {
		return nil, err
	}
	return &chainsResponse.ChainIds, nil
}

func (n *NodeConfig) QueryChainsInfo(chainIDs []string) ([]*zctypes.ChainInfo, error) {
	queryParams := url.Values{}
	for _, chainId := range chainIDs {
		queryParams.Add("chain_ids", chainId)
	}

	bz, err := n.QueryGRPCGateway("/babylon/zoneconcierge/v1/chains_info", queryParams)
	require.NoError(n.t, err)
	var resp zctypes.QueryChainsInfoResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}
	return resp.ChainsInfo, nil
}

func (n *NodeConfig) QueryCurrentEpoch() (uint64, error) {
	bz, err := n.QueryGRPCGateway("/babylon/epoching/v1/current_epoch", url.Values{})
	require.NoError(n.t, err)
	var epochResponse etypes.QueryCurrentEpochResponse
	if err := util.Cdc.UnmarshalJSON(bz, &epochResponse); err != nil {
		return 0, err
	}
	return epochResponse.CurrentEpoch, nil
}

func (n *NodeConfig) QueryLightClientHeightEpochEnd(epoch uint64) (uint64, error) {
	monitorPath := fmt.Sprintf("/babylon/monitor/v1/epochs/%d", epoch)
	bz, err := n.QueryGRPCGateway(monitorPath, url.Values{})
	require.NoError(n.t, err)
	var mResponse mtypes.QueryEndedEpochBtcHeightResponse
	if err := util.Cdc.UnmarshalJSON(bz, &mResponse); err != nil {
		return 0, err
	}
	return mResponse.BtcLightClientHeight, nil
}

func (n *NodeConfig) QueryLightClientHeightCheckpointReported(ckptHash []byte) (uint64, error) {
	monitorPath := fmt.Sprintf("/babylon/monitor/v1/checkpoints/%x", ckptHash)
	bz, err := n.QueryGRPCGateway(monitorPath, url.Values{})
	require.NoError(n.t, err)
	var mResponse mtypes.QueryReportedCheckpointBtcHeightResponse
	if err := util.Cdc.UnmarshalJSON(bz, &mResponse); err != nil {
		return 0, err
	}
	return mResponse.BtcLightClientHeight, nil
}

func (n *NodeConfig) QueryLatestWasmCodeID() uint64 {
	path := "/cosmwasm/wasm/v1/code"

	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var response wasmtypes.QueryCodesResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err)
	if len(response.CodeInfos) == 0 {
		return 0
	}
	return response.CodeInfos[len(response.CodeInfos)-1].CodeID
}

func (n *NodeConfig) QueryContractsFromId(codeId int) ([]string, error) {
	path := fmt.Sprintf("/cosmwasm/wasm/v1/code/%d/contracts", codeId)
	bz, err := n.QueryGRPCGateway(path, url.Values{})

	require.NoError(n.t, err)

	var contractsResponse wasmtypes.QueryContractsByCodeResponse
	if err := util.Cdc.UnmarshalJSON(bz, &contractsResponse); err != nil {
		return nil, err
	}

	return contractsResponse.Contracts, nil
}

func (n *NodeConfig) QueryWasmSmart(contract string, msg string, result any) error {
	// base64-encode the msg
	encodedMsg := base64.StdEncoding.EncodeToString([]byte(msg))
	path := fmt.Sprintf("/cosmwasm/wasm/v1/contract/%s/smart/%s", contract, encodedMsg)

	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return err
	}

	var response wasmtypes.QuerySmartContractStateResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	if err != nil {
		return err
	}

	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeConfig) QueryWasmSmartObject(contract string, msg string) (resultObject map[string]interface{}, err error) {
	err = n.QueryWasmSmart(contract, msg, &resultObject)
	if err != nil {
		return nil, err
	}
	return resultObject, nil
}
