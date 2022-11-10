package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
)

// Default constants
const (
	chainID        = ""
	keyringBackend = "os"
	output         = "text"
	node           = "tcp://localhost:26657"
	broadcastMode  = "sync"
	fromName       = ""
)

// adapted from https://github.com/cosmos/cosmos-sdk/blob/6d32debf1aca4b7f1ed1429d87be1d02c315f02d/client/config/config.go
type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	FromName       string `mapstructure:"from-name" json:"from-name"`
}

// defaultClientConfig returns the reference to ClientConfig with default values.
func defaultClientConfig() *ClientConfig {
	return &ClientConfig{chainID, keyringBackend, output, node, broadcastMode, fromName}
}

func (c *ClientConfig) SetChainID(chainID string) {
	c.ChainID = chainID
}

func (c *ClientConfig) SetKeyringBackend(keyringBackend string) {
	c.KeyringBackend = keyringBackend
}

func (c *ClientConfig) SetOutput(output string) {
	c.Output = output
}

func (c *ClientConfig) SetNode(node string) {
	c.Node = node
}

func (c *ClientConfig) SetBroadcastMode(broadcastMode string) {
	c.BroadcastMode = broadcastMode
}

func (c *ClientConfig) SetFromName(fromName string) {
	c.FromName = fromName
}

func CreateClientConfig(chainID string, backend string, homePath string, fromName string) (*ClientConfig, error) {
	cliConf := &ClientConfig{
		ChainID:        chainID,
		KeyringBackend: backend,
		Output:         "text",                  // default value from config.ClientConfig
		Node:           "tcp://localhost:26657", // default value from config.ClientConfig
		BroadcastMode:  "sync",                  // default value from config.ClientConfig
		FromName:       fromName,
	}
	err := saveClientConfig(homePath, cliConf)
	if err != nil {
		return nil, err
	}

	return cliConf, err
}

// ReadFromClientConfig reads values from client.toml file and updates them in client Context
func ReadFromClientConfig(ctx client.Context) (client.Context, error) {
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")
	conf := defaultClientConfig()

	// if config.toml file does not exist we create it and write default ClientConfig values into it.
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := ensureConfigPath(configPath); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %v", err)
		}

		if ctx.ChainID != "" {
			conf.ChainID = ctx.ChainID // chain-id will be written to the client.toml while initiating the chain.
		}

		if err := writeConfigToFile(configFilePath, conf); err != nil {
			return ctx, fmt.Errorf("could not write client config to the file: %v", err)
		}
	}

	conf, err := getClientConfig(configPath, ctx.Viper)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client config: %v", err)
	}
	// we need to update KeyringDir field on Client Context first cause it is used in NewKeyringFromBackend
	ctx = ctx.WithOutputFormat(conf.Output).
		WithChainID(conf.ChainID).
		WithKeyringDir(ctx.HomeDir)

	keyring, err := client.NewKeyringFromBackend(ctx, conf.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get key ring: %v", err)
	}

	ctx = ctx.WithKeyring(keyring)

	// https://github.com/cosmos/cosmos-sdk/issues/8986
	client, err := client.NewClientFromNode(conf.Node)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client from nodeURI: %v", err)
	}

	ctx = ctx.WithNodeURI(conf.Node).
		WithClient(client).
		WithBroadcastMode(conf.BroadcastMode)

	if ctx.FromName == "" {
		ctx = ctx.WithFromName(conf.FromName)
	}

	return ctx, nil
}

func saveClientConfig(homePath string, cliConf *ClientConfig) error {
	var err error
	configPath := filepath.Join(homePath, "config")
	configFilePath := filepath.Join(configPath, "client.toml")
	if err = ensureConfigPath(configPath); err != nil {
		return fmt.Errorf("couldn't make client config: %v", err)
	}

	if err = writeConfigToFile(configFilePath, cliConf); err != nil {
		return fmt.Errorf("could not write client config to the file: %v", err)
	}

	return nil
}
