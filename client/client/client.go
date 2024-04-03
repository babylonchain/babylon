package client

import (
	"context"
	"time"

	bbn "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/client/config"
	"github.com/babylonchain/babylon/client/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"go.uber.org/zap"
)

type Client struct {
	*query.QueryClient

	provider *cosmos.CosmosProvider
	timeout  time.Duration
	logger   *zap.Logger
	cfg      *config.BabylonConfig
}

func New(cfg *config.BabylonConfig, logger *zap.Logger) (*Client, error) {
	var (
		zapLogger *zap.Logger
		err       error
	)

	// ensure cfg is valid
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// use the existing logger or create a new one if not given
	zapLogger = logger
	if zapLogger == nil {
		zapLogger, err = newRootLogger("console", true)
		if err != nil {
			return nil, err
		}
	}

	provider, err := cfg.ToCosmosProviderConfig().NewProvider(
		zapLogger,
		"", // TODO: set home path
		true,
		"babylon",
	)
	if err != nil {
		return nil, err
	}

	cp := provider.(*cosmos.CosmosProvider)
	cp.PCfg.KeyDirectory = cfg.KeyDirectory

	// Create tmp Babylon app to retrieve and register codecs
	// Need to override this manually as otherwise option from config is ignored
	encCfg := bbn.GetEncodingConfig()
	cp.Cdc = cosmos.Codec{
		InterfaceRegistry: encCfg.InterfaceRegistry,
		Marshaler:         encCfg.Codec,
		TxConfig:          encCfg.TxConfig,
		Amino:             encCfg.Amino,
	}

	// initialise Cosmos provider
	// NOTE: this will create a RPC client. The RPC client will be used for
	// submitting txs and making ad hoc queries. It won't create WebSocket
	// connection with Babylon node
	err = cp.Init(context.Background())
	if err != nil {
		return nil, err
	}

	// create a queryClient so that the Client inherits all query functions
	// TODO: merge this RPC client with the one in `cp` after Cosmos side
	// finishes the migration to new RPC client
	// see https://github.com/strangelove-ventures/cometbft-client
	c, err := rpchttp.NewWithTimeout(cp.PCfg.RPCAddr, "/websocket", uint(cfg.Timeout.Seconds()))
	if err != nil {
		return nil, err
	}
	queryClient, err := query.NewWithClient(c, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	return &Client{
		queryClient,
		cp,
		cfg.Timeout,
		zapLogger,
		cfg,
	}, nil
}

func (c *Client) GetConfig() *config.BabylonConfig {
	return c.cfg
}
