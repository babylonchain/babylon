package privval

import (
	"bytes"
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/tendermint/tendermint/privval"
	"io/ioutil"
	"time"

	"github.com/gogo/protobuf/proto"

	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/libs/protoio"
	"github.com/tendermint/tendermint/libs/tempfile"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
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

// LoadFilePV loads a FilePV from the filePaths.  The FilePV handles double
// signing prevention by persisting data to the stateFilePath.  If either file path
// does not exist, the program will exit.
func LoadFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	return loadFilePV(keyFilePath, stateFilePath, true)
}

// LoadFilePVEmptyState loads a FilePV from the given keyFilePath, with an empty LastSignState.
// If the keyFilePath does not exist, the program will exit.
func LoadFilePVEmptyState(keyFilePath, stateFilePath string) *WrappedFilePV {
	return loadFilePV(keyFilePath, stateFilePath, false)
}

// If loadState is true, we load from the stateFilePath. Otherwise, we use an empty LastSignState.
func loadFilePV(keyFilePath, stateFilePath string, loadState bool) *WrappedFilePV {
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

// LoadOrGenFilePV loads a FilePV from the given filePaths
// or else generates a new one and saves it to the filePaths.
func LoadOrGenFilePV(keyFilePath, stateFilePath string) *WrappedFilePV {
	var pv *WrappedFilePV
	if tmos.FileExists(keyFilePath) {
		pv = LoadFilePV(keyFilePath, stateFilePath)
	} else {
		pv = GenWrappedFilePV(keyFilePath, stateFilePath)
		pv.Save()
	}
	return pv
}

// GetAddress returns the address of the validator.
// Implements PrivValidator.
func (pv *WrappedFilePV) GetAddress() types.Address {
	return pv.Key.Address
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

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements PrivValidator.
func (pv *WrappedFilePV) SignVote(chainID string, vote *tmproto.Vote) error {
	if err := pv.signVote(chainID, vote); err != nil {
		return fmt.Errorf("error signing vote: %v", err)
	}
	return nil
}

// SignProposal signs a canonical representation of the proposal, along with
// the chainID. Implements PrivValidator.
func (pv *WrappedFilePV) SignProposal(chainID string, proposal *tmproto.Proposal) error {
	if err := pv.signProposal(chainID, proposal); err != nil {
		return fmt.Errorf("error signing proposal: %v", err)
	}
	return nil
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

//------------------------------------------------------------------------------------

// signVote checks if the vote is good to sign and sets the vote signature.
// It may need to set the timestamp as well if the vote is otherwise the same as
// a previously signed vote (ie. we crashed after signing but before the vote hit the WAL).
func (pv *WrappedFilePV) signVote(chainID string, vote *tmproto.Vote) error {
	height, round, step := vote.Height, vote.Round, voteToStep(vote)

	lss := pv.LastSignState

	sameHRS, err := lss.CheckHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.VoteSignBytes(chainID, vote)

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, lss.SignBytes) {
			vote.Signature = lss.Signature
		} else if timestamp, ok := checkVotesOnlyDifferByTimestamp(lss.SignBytes, signBytes); ok {
			vote.Timestamp = timestamp
			vote.Signature = lss.Signature
		} else {
			err = fmt.Errorf("conflicting data")
		}
		return err
	}

	// It passed the checks. Sign the vote
	sig, err := pv.Key.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	pv.saveSigned(height, round, step, signBytes, sig)
	vote.Signature = sig
	return nil
}

// signProposal checks if the proposal is good to sign and sets the proposal signature.
// It may need to set the timestamp as well if the proposal is otherwise the same as
// a previously signed proposal ie. we crashed after signing but before the proposal hit the WAL).
func (pv *WrappedFilePV) signProposal(chainID string, proposal *tmproto.Proposal) error {
	height, round, step := proposal.Height, proposal.Round, stepPropose

	lss := pv.LastSignState

	sameHRS, err := lss.CheckHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.ProposalSignBytes(chainID, proposal)

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, lss.SignBytes) {
			proposal.Signature = lss.Signature
		} else if timestamp, ok := checkProposalsOnlyDifferByTimestamp(lss.SignBytes, signBytes); ok {
			proposal.Timestamp = timestamp
			proposal.Signature = lss.Signature
		} else {
			err = fmt.Errorf("conflicting data")
		}
		return err
	}

	// It passed the checks. Sign the proposal
	sig, err := pv.Key.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	pv.saveSigned(height, round, step, signBytes, sig)
	proposal.Signature = sig
	return nil
}

// Persist height/round/step and signature
func (pv *WrappedFilePV) saveSigned(height int64, round int32, step int8,
	signBytes []byte, sig []byte) {

	pv.LastSignState.Height = height
	pv.LastSignState.Round = round
	pv.LastSignState.Step = step
	pv.LastSignState.Signature = sig
	pv.LastSignState.SignBytes = signBytes
	pv.LastSignState.Save()
}

//-----------------------------------------------------------------------------------------

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the votes is their timestamp.
func checkVotesOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastVote, newVote tmproto.CanonicalVote
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastVote); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into vote: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newVote); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into vote: %v", err))
	}

	lastTime := lastVote.Timestamp
	// set the times to the same value and check equality
	now := tmtime.Now()
	lastVote.Timestamp = now
	newVote.Timestamp = now

	return lastTime, proto.Equal(&newVote, &lastVote)
}

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the proposals is their timestamp
func checkProposalsOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastProposal, newProposal tmproto.CanonicalProposal
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastProposal); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into proposal: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newProposal); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into proposal: %v", err))
	}

	lastTime := lastProposal.Timestamp
	// set the times to the same value and check equality
	now := tmtime.Now()
	lastProposal.Timestamp = now
	newProposal.Timestamp = now

	return lastTime, proto.Equal(&newProposal, &lastProposal)
}
