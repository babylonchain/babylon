package initialization

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type ChainMeta struct {
	DataDir string `json:"dataDir"`
	Id      string `json:"id"`
}

type Node struct {
	Name          string `json:"name"`
	ConfigDir     string `json:"configDir"`
	Mnemonic      string `json:"mnemonic"`
	PublicAddress string `json:"publicAddress"`
	WalletName    string `json:"walletName"`
	SecretKey     cryptotypes.PrivKey
	PublicKey     []byte `json:"publicKey"`
	PeerId        string `json:"peerId"`
	IsValidator   bool   `json:"isValidator"`
}

type Chain struct {
	ChainMeta ChainMeta `json:"chainMeta"`
	Nodes     []*Node   `json:"validators"`
}

func (c *ChainMeta) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.Id)
}
