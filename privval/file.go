package privval

import (
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/libs/tempfile"
	"github.com/tendermint/tendermint/privval"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	"io/ioutil"
	"os"
	"path/filepath"
)

// copied from github.com/tendermint/tendermint/privval/file.go"
const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

// copied from github.com/tendermint/tendermint/privval/file.go"
func voteToStep(vote *tmproto.Vote) int8 {
	switch vote.Type {
	case tmproto.PrevoteType:
		return stepPrevote
	case tmproto.PrecommitType:
		return stepPrecommit
	default:
		panic(fmt.Sprintf("Unknown vote type: %v", vote.Type))
	}
}

// WrappedFilePVKey wraps FilePVKey with BLS keys.
type WrappedFilePVKey struct {
	AccAddress string              `json:"acc_address"`
	Address    types.Address       `json:"address"`
	PubKey     tmcrypto.PubKey     `json:"pub_key"`
	PrivKey    tmcrypto.PrivKey    `json:"priv_key"`
	BlsPubKey  bls12381.PublicKey  `json:"bls_pub_key"`
	BlsPrivKey bls12381.PrivateKey `json:"bls_priv_key"`

	filePath string
}

// Save persists the FilePVKey to its filePath.
func (pvKey WrappedFilePVKey) Save() {
	outFile := pvKey.filePath
	if outFile == "" {
		panic("cannot save PrivValidator key: filePath not set")
	}

	jsonBytes, err := tmjson.MarshalIndent(pvKey, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := tempfile.WriteFileAtomic(outFile, jsonBytes, 0600); err != nil {
		panic(err)
	}
}

//-------------------------------------------------------------------------------

// WrappedFilePV wraps FilePV with WrappedFilePVKey.
type WrappedFilePV struct {
	Key           WrappedFilePVKey
	LastSignState privval.FilePVLastSignState
}

// NewWrappedFilePV wraps FilePV
func NewWrappedFilePV(privKey tmcrypto.PrivKey, blsPrivKey bls12381.PrivateKey, keyFilePath, stateFilePath string) *WrappedFilePV {
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
	keyJSONBytes, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		tmos.Exit(err.Error())
	}
	pvKey := WrappedFilePVKey{}
	err = tmjson.Unmarshal(keyJSONBytes, &pvKey)
	if err != nil {
		tmos.Exit(fmt.Sprintf("Error reading PrivValidator key from %v: %v\n", keyFilePath, err))
	}

	// overwrite pubkey and address for convenience
	pvKey.PubKey = pvKey.PrivKey.PubKey()
	pvKey.Address = pvKey.PubKey.Address()
	pvKey.BlsPubKey = pvKey.BlsPrivKey.PubKey()
	pvKey.filePath = keyFilePath

	pvState := privval.FilePVLastSignState{}

	if loadState {
		stateJSONBytes, err := ioutil.ReadFile(stateFilePath)
		if err != nil {
			tmos.Exit(err.Error())
		}
		err = tmjson.Unmarshal(stateJSONBytes, &pvState)
		if err != nil {
			tmos.Exit(fmt.Sprintf("Error reading PrivValidator state from %v: %v\n", stateFilePath, err))
		}
	}

	// adding path is not needed
	//pvState.filePath = stateFilePath

	return &WrappedFilePV{
		Key:           pvKey,
		LastSignState: pvState,
	}
}

// LoadOrGenWrappedFilePV loads a FilePV from the given filePaths
// or else generates a new one and saves it to the filePaths.
func LoadOrGenWrappedFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	var pv *WrappedFilePV
	if tmos.FileExists(keyFilePath) {
		pv = LoadWrappedFilePV(keyFilePath, stateFilePath)
	} else {
		pv = GenWrappedFilePV(keyFilePath, stateFilePath)
		pv.Save()
	}
	return pv
}

// GetAddress returns the address of the validator.
// Implements PrivValidator.
func (pv *WrappedFilePV) GetAddress() sdk.ValAddress {
	if pv.Key.AccAddress == "" {
		return sdk.ValAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(pv.Key.AccAddress)
	if err != nil {
		panic(err)
	}
	return sdk.ValAddress(addr)
}

func (pv *WrappedFilePV) SetAccAddress(addr sdk.AccAddress) {
	pv.Key.AccAddress = addr.String()
	pv.Key.Save()
}

// GetPubKey returns the public key of the validator.
// Implements PrivValidator.
func (pv *WrappedFilePV) GetPubKey() (tmcrypto.PubKey, error) {
	return pv.Key.PubKey, nil
}

func (pv *WrappedFilePV) GetValPrivKey() tmcrypto.PrivKey {
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
