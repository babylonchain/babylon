package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	sdkmath "cosmossdk.io/math"

	tmabcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/test/e2e/util"
	blc "github.com/babylonchain/babylon/x/btclightclient/types"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	mtypes "github.com/babylonchain/babylon/x/monitor/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
)

func (n *NodeConfig) QueryGRPCGateway(path string, parameters ...string) ([]byte, error) {
	if len(parameters)%2 != 0 {
		return nil, fmt.Errorf("invalid number of parameters, must follow the format of key + value")
	}

	// add the URL for the given validator ID, and pre-pend to to path.
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

		if len(parameters) > 0 {
			q := req.URL.Query()
			for i := 0; i < len(parameters); i += 2 {
				q.Add(parameters[i], parameters[i+1])
			}
			req.URL.RawQuery = q.Encode()
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

// QueryBalancer returns balances at the address.
func (n *NodeConfig) QueryBalances(address string) (sdk.Coins, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s", address)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.Coins{}, err
	}
	return balancesResp.GetBalances(), nil
}

func (n *NodeConfig) QuerySupplyOf(denom string) (sdkmath.Int, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/supply/%s", denom)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var supplyResp banktypes.QuerySupplyOfResponse
	if err := util.Cdc.UnmarshalJSON(bz, &supplyResp); err != nil {
		return sdk.NewInt(0), err
	}
	return supplyResp.Amount.Amount, nil
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

func (n *NodeConfig) QueryCheckpointForEpoch(epoch uint64) (*ct.RawCheckpointWithMeta, error) {
	path := fmt.Sprintf("babylon/checkpointing/v1/raw_checkpoint/%d", epoch)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var checkpointingResponse ct.QueryRawCheckpointResponse
	if err := util.Cdc.UnmarshalJSON(bz, &checkpointingResponse); err != nil {
		return nil, err
	}

	return checkpointingResponse.RawCheckpoint, nil
}

func (n *NodeConfig) QueryBtcBaseHeader() (*blc.BTCHeaderInfo, error) {
	bz, err := n.QueryGRPCGateway("babylon/btclightclient/v1/baseheader")
	require.NoError(n.t, err)

	var blcResponse blc.QueryBaseHeaderResponse
	if err := util.Cdc.UnmarshalJSON(bz, &blcResponse); err != nil {
		return nil, err
	}

	return blcResponse.Header, nil
}

func (n *NodeConfig) QueryTip() (*blc.BTCHeaderInfo, error) {
	bz, err := n.QueryGRPCGateway("babylon/btclightclient/v1/tip")
	require.NoError(n.t, err)

	var blcResponse blc.QueryTipResponse
	if err := util.Cdc.UnmarshalJSON(bz, &blcResponse); err != nil {
		return nil, err
	}

	return blcResponse.Header, nil
}

func (n *NodeConfig) QueryFinalizedChainInfo(chainId string) (*zctypes.QueryFinalizedChainInfoResponse, error) {
	finalizedPath := fmt.Sprintf("babylon/zoneconcierge/v1/finalized_chain_info/%s", chainId)
	bz, err := n.QueryGRPCGateway(finalizedPath)
	require.NoError(n.t, err)

	var finalizedResponse zctypes.QueryFinalizedChainInfoResponse
	if err := util.Cdc.UnmarshalJSON(bz, &finalizedResponse); err != nil {
		return nil, err
	}

	return &finalizedResponse, nil
}

func (n *NodeConfig) QueryCheckpointChains() (*[]string, error) {
	bz, err := n.QueryGRPCGateway("babylon/zoneconcierge/v1/chains")
	require.NoError(n.t, err)
	var chainsResponse zctypes.QueryChainListResponse
	if err := util.Cdc.UnmarshalJSON(bz, &chainsResponse); err != nil {
		return nil, err
	}
	return &chainsResponse.ChainIds, nil
}

func (n *NodeConfig) QueryCheckpointChainInfo(chainId string) (*zctypes.ChainInfo, error) {
	infoPath := fmt.Sprintf("/babylon/zoneconcierge/v1/chain_info/%s", chainId)
	bz, err := n.QueryGRPCGateway(infoPath)
	require.NoError(n.t, err)
	var infoResponse zctypes.QueryChainInfoResponse
	if err := util.Cdc.UnmarshalJSON(bz, &infoResponse); err != nil {
		return nil, err
	}
	return infoResponse.ChainInfo, nil
}

func (n *NodeConfig) QueryCurrentEpoch() (uint64, error) {
	bz, err := n.QueryGRPCGateway("/babylon/epoching/v1/current_epoch")
	require.NoError(n.t, err)
	var epochResponse etypes.QueryCurrentEpochResponse
	if err := util.Cdc.UnmarshalJSON(bz, &epochResponse); err != nil {
		return 0, err
	}
	return epochResponse.CurrentEpoch, nil
}

func (n *NodeConfig) QueryLightClientHeightEpochEnd(epoch uint64) (uint64, error) {
	monitorPath := fmt.Sprintf("/babylon/monitor/v1/epochs/%d", epoch)
	bz, err := n.QueryGRPCGateway(monitorPath)
	require.NoError(n.t, err)
	var mResponse mtypes.QueryEndedEpochBtcHeightResponse
	if err := util.Cdc.UnmarshalJSON(bz, &mResponse); err != nil {
		return 0, err
	}
	return mResponse.BtcLightClientHeight, nil
}

func (n *NodeConfig) QueryLightClientHeightCheckpointReported(ckptHash []byte) (uint64, error) {
	monitorPath := fmt.Sprintf("/babylon/monitor/v1/checkpoints/%x", ckptHash)
	bz, err := n.QueryGRPCGateway(monitorPath)
	require.NoError(n.t, err)
	var mResponse mtypes.QueryReportedCheckpointBtcHeightResponse
	if err := util.Cdc.UnmarshalJSON(bz, &mResponse); err != nil {
		return 0, err
	}
	return mResponse.BtcLightClientHeight, nil
}
