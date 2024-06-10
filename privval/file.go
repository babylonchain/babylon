package privval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/libs/tempfile"
	"github.com/cometbft/cometbft/privval"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

// copied from github.com/cometbft/cometbft/privval/file.go"
//
//nolint:unused
const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

// copied from github.com/cometbft/cometbft/privval/file.go"
//
//nolint:unused
func voteToStep(vote *cmtproto.Vote) int8 {
	switch vote.Type {
	case cmtproto.PrevoteType:
		return stepPrevote
	case cmtproto.PrecommitType:
		return stepPrecommit
	default:
		panic(fmt.Sprintf("Unknown vote type: %v", vote.Type))
	}
}

// WrappedFilePVKey wraps FilePVKey with BLS keys.
type WrappedFilePVKey struct {
	DelegatorAddress string              `json:"acc_address"`
	Address          types.Address       `json:"address"`
	PubKey           cmtcrypto.PubKey    `json:"pub_key"`
	PrivKey          cmtcrypto.PrivKey   `json:"priv_key"`
	BlsPubKey        bls12381.PublicKey  `json:"bls_pub_key"`
	BlsPrivKey       bls12381.PrivateKey `json:"bls_priv_key"`

	filePath string
}

// Save persists the FilePVKey to its filePath.
func (pvKey WrappedFilePVKey) Save() {
	outFile := pvKey.filePath
	if outFile == "" {
		panic("cannot save PrivValidator key: filePath not set")
	}

	jsonBytes, err := cmtjson.MarshalIndent(pvKey, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := tempfile.WriteFileAtomic(outFile, jsonBytes, 0600); err != nil {
		panic(err)
	}
}

// -------------------------------------------------------------------------------

// WrappedFilePV wraps FilePV with WrappedFilePVKey.
type WrappedFilePV struct {
	Key           WrappedFilePVKey
	LastSignState privval.FilePVLastSignState
}

// NewWrappedFilePV wraps FilePV
func NewWrappedFilePV(privKey cmtcrypto.PrivKey, blsPrivKey bls12381.PrivateKey, keyFilePath, stateFilePath string) *WrappedFilePV {
	filePV := privval.NewFilePV(privKey, keyFilePath, stateFilePath)
	return &WrappedFilePV{
		Key: WrappedFilePVKey{
			Address:    privKey.PubKey().Address(),
			PubKey:     privKey.PubKey(),
			PrivKey:    privKey,
			BlsPubKey:  blsPrivKey.PubKey(),
			BlsPrivKey: blsPrivKey,
			filePath:   keyFilePath,
		},
		LastSignState: filePV.LastSignState,
	}
}

// GenWrappedFilePV generates a new validator with randomly generated private key
// and sets the filePaths, but does not call Save().
func GenWrappedFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	return NewWrappedFilePV(ed25519.GenPrivKey(), bls12381.GenPrivKey(), keyFilePath, stateFilePath)
}

// LoadWrappedFilePV loads a FilePV from the filePaths.  The FilePV handles double
// signing prevention by persisting data to the stateFilePath.  If either file path
// does not exist, the program will exit.
func LoadWrappedFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	return loadWrappedFilePV(keyFilePath, stateFilePath, true)
}

// LoadWrappedFilePVEmptyState loads a FilePV from the given keyFilePath, with an empty LastSignState.
// If the keyFilePath does not exist, the program will exit.
func LoadWrappedFilePVEmptyState(keyFilePath, stateFilePath string) *WrappedFilePV {
	return loadWrappedFilePV(keyFilePath, stateFilePath, false)
}

// If loadState is true, we load from the stateFilePath. Otherwise, we use an empty LastSignState.
func loadWrappedFilePV(keyFilePath, stateFilePath string, loadState bool) *WrappedFilePV {
	keyFilePath = filepath.Clean(keyFilePath)
	keyJSONBytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		cmtos.Exit(err.Error())
	}
	pvKey := WrappedFilePVKey{}
	err = cmtjson.Unmarshal(keyJSONBytes, &pvKey)
	if err != nil {
		cmtos.Exit(fmt.Sprintf("Error reading PrivValidator key from %v: %v\n", keyFilePath, err))
	}

	// overwrite pubkey and address for convenience
	pvKey.PubKey = pvKey.PrivKey.PubKey()
	pvKey.Address = pvKey.PubKey.Address()
	pvKey.BlsPubKey = pvKey.BlsPrivKey.PubKey()
	pvKey.filePath = keyFilePath

	pvState := privval.FilePVLastSignState{}

	if loadState {
		stateFilePath := filepath.Clean(stateFilePath)
		stateJSONBytes, err := os.ReadFile(stateFilePath)
		if err != nil {
			cmtos.Exit(err.Error())
		}
		err = cmtjson.Unmarshal(stateJSONBytes, &pvState)
		if err != nil {
			cmtos.Exit(fmt.Sprintf("Error reading PrivValidator state from %v: %v\n", stateFilePath, err))
		}
	}

	// adding path is not needed
	// pvState.filePath = stateFilePath

	return &WrappedFilePV{
		Key:           pvKey,
		LastSignState: pvState,
	}
}

// LoadOrGenWrappedFilePV loads a FilePV from the given filePaths
// or else generates a new one and saves it to the filePaths.
func LoadOrGenWrappedFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	var pv *WrappedFilePV
	if cmtos.FileExists(keyFilePath) {
		pv = LoadWrappedFilePV(keyFilePath, stateFilePath)
	} else {
		pv = GenWrappedFilePV(keyFilePath, stateFilePath)
		pv.Save()
	}
	return pv
}

// ExportGenBls writes a {address, bls_pub_key, pop, and pub_key} into a json file
func (pv *WrappedFilePV) ExportGenBls(filePath string) (outputFileName string, err error) {
	if !cmtos.FileExists(filePath) {
		return outputFileName, errors.New("export file path does not exist")
	}

	valAddress := pv.GetAddress()
	if valAddress.Empty() {
		return outputFileName, errors.New("validator address should not be empty")
	}

	validatorKey, err := NewValidatorKeys(pv.GetValPrivKey(), pv.GetBlsPrivKey())
	if err != nil {
		return outputFileName, err
	}

	pubkey, err := codec.FromCmtPubKeyInterface(validatorKey.ValPubkey)
	if err != nil {
		return outputFileName, err
	}

	genbls, err := checkpointingtypes.NewGenesisKey(valAddress, &validatorKey.BlsPubkey, validatorKey.PoP, pubkey)
	if err != nil {
		return outputFileName, err
	}

	jsonBytes, err := cmtjson.MarshalIndent(genbls, "", "  ")
	if err != nil {
		return outputFileName, err
	}

	outputFileName = filepath.Join(filePath, fmt.Sprintf("gen-bls-%s.json", valAddress.String()))
	err = tempfile.WriteFileAtomic(outputFileName, jsonBytes, 0600)
	return outputFileName, err
}

// GetAddress returns the delegator address of the validator.
// Implements PrivValidator.
func (pv *WrappedFilePV) GetAddress() sdk.ValAddress {
	if pv.Key.DelegatorAddress == "" {
		return sdk.ValAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(pv.Key.DelegatorAddress)
	if err != nil {
		cmtos.Exit(err.Error())
	}
	return sdk.ValAddress(addr)
}

func (pv *WrappedFilePV) SetAccAddress(addr sdk.AccAddress) {
	pv.Key.DelegatorAddress = addr.String()
	pv.Key.Save()
}

// GetPubKey returns the public key of the validator.
// Implements PrivValidator.
func (pv *WrappedFilePV) GetPubKey() (cmtcrypto.PubKey, error) {
	return pv.Key.PubKey, nil
}

func (pv *WrappedFilePV) GetValPrivKey() cmtcrypto.PrivKey {
	return pv.Key.PrivKey
}

func (pv *WrappedFilePV) GetBlsPrivKey() bls12381.PrivateKey {
	return pv.Key.BlsPrivKey
}

func (pv *WrappedFilePV) SignMsgWithBls(msg []byte) (bls12381.Signature, error) {
	blsPrivKey := pv.GetBlsPrivKey()
	if blsPrivKey == nil {
		return nil, checkpointingtypes.ErrBlsPrivKeyDoesNotExist
	}
	return bls12381.Sign(blsPrivKey, msg), nil
}

func (pv *WrappedFilePV) GetBlsPubkey() (bls12381.PublicKey, error) {
	blsPrivKey := pv.GetBlsPrivKey()
	if blsPrivKey == nil {
		return nil, checkpointingtypes.ErrBlsPrivKeyDoesNotExist
	}
	return blsPrivKey.PubKey(), nil
}

func (pv *WrappedFilePV) GetValidatorPubkey() (cmtcrypto.PubKey, error) {
	return pv.GetPubKey()
}

// Save persists the FilePV to disk.
func (pv *WrappedFilePV) Save() {
	pv.Key.Save()
	pv.LastSignState.Save()
}

// Reset resets all fields in the FilePV.
// NOTE: Unsafe!
func (pv *WrappedFilePV) Reset() {
	var sig []byte
	pv.LastSignState.Height = 0
	pv.LastSignState.Round = 0
	pv.LastSignState.Step = 0
	pv.LastSignState.Signature = sig
	pv.LastSignState.SignBytes = nil
	pv.Save()
}

// Clean removes PVKey file and PVState file
func (pv *WrappedFilePV) Clean(keyFilePath, stateFilePath string) {
	_ = os.RemoveAll(filepath.Dir(keyFilePath))
	_ = os.RemoveAll(filepath.Dir(stateFilePath))
}

// String returns a string representation of the FilePV.
func (pv *WrappedFilePV) String() string {
	return fmt.Sprintf(
		"PrivValidator{%v LH:%v, LR:%v, LS:%v}",
		pv.GetAddress(),
		pv.LastSignState.Height,
		pv.LastSignState.Round,
		pv.LastSignState.Step,
	)
}
