package datagen

import (
	"fmt"
	"path/filepath"

	cfg "github.com/cometbft/cometbft/config"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
)

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(config *cfg.Config, addr sdk.AccAddress) (string, *privval.ValidatorKeys, error) {
	return InitializeNodeValidatorFilesFromMnemonic(config, "", addr)
}

func InitializeNodeValidatorFilesFromMnemonic(config *cfg.Config, mnemonic string, addr sdk.AccAddress) (nodeID string, valKeys *privval.ValidatorKeys, err error) {
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, fmt.Errorf("invalid mnemonic")
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}

	nodeID = string(nodeKey.ID())

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := cmtos.EnsureDir(filepath.Dir(pvKeyFile), 0777); err != nil {
		return "", nil, err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := cmtos.EnsureDir(filepath.Dir(pvStateFile), 0777); err != nil {
		return "", nil, err
	}

	var filePV *privval.WrappedFilePV
	if len(mnemonic) == 0 {
		filePV = privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)
	} else {
		privKey := cmted25519.GenPrivKeyFromSecret([]byte(mnemonic))
		blsPrivKey := bls12381.GenPrivKeyFromSecret([]byte(mnemonic))
		filePV = privval.NewWrappedFilePV(privKey, blsPrivKey, pvKeyFile, pvStateFile)
	}
	filePV.SetAccAddress(addr)

	valPrivkey := filePV.GetValPrivKey()
	blsPrivkey := filePV.GetBlsPrivKey()
	valKeys, err = privval.NewValidatorKeys(valPrivkey, blsPrivkey)
	if err != nil {
		return "", nil, err
	}

	return nodeID, valKeys, nil
}
