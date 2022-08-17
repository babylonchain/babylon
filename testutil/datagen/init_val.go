package datagen

import (
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/go-bip39"
	cfg "github.com/tendermint/tendermint/config"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/p2p"
	"path/filepath"
)

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(config *cfg.Config) (string, *privval.ValidatorKeys, error) {
	return InitializeNodeValidatorFilesFromMnemonic(config, "")
}

func InitializeNodeValidatorFilesFromMnemonic(config *cfg.Config, mnemonic string) (nodeID string, valKeys *privval.ValidatorKeys, err error) {
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, fmt.Errorf("invalid mnemonic")
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}

	nodeID = string(nodeKey.ID())

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0777); err != nil {
		return "", nil, err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := tmos.EnsureDir(filepath.Dir(pvStateFile), 0777); err != nil {
		return "", nil, err
	}

	var filePV *privval.WrappedFilePV
	if len(mnemonic) == 0 {
		filePV = privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)
	} else {
		privKey := tmed25519.GenPrivKeyFromSecret([]byte(mnemonic))
		blsPrivKey := bls12381.GenPrivKeyFromSecret([]byte(mnemonic))
		filePV = privval.NewWrappedFilePV(privKey, blsPrivKey, pvKeyFile, pvStateFile)
	}

	valPrivkey := filePV.GetValPrivKey()
	blsPrivkey := filePV.GetBlsPrivKey()
	valKeys, err = privval.NewValidatorKeys(valPrivkey, blsPrivkey)
	if err != nil {
		return "", nil, err
	}

	return nodeID, valKeys, nil
}
