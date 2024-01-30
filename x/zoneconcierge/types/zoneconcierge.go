package types

import (
	"bytes"
	"fmt"

	"cosmossdk.io/store/rootmulti"
	"github.com/cometbft/cometbft/crypto/merkle"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

// VerifyStore verifies whether a KV pair is committed to the Merkle root, with the assistance of a Merkle proof
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/store/rootmulti/proof_test.go)
func VerifyStore(root []byte, moduleStoreKey string, key []byte, value []byte, proof *cmtcrypto.ProofOps) error {
	prt := rootmulti.DefaultProofRuntime()

	keypath := merkle.KeyPath{}
	keypath = keypath.AppendKey([]byte(moduleStoreKey), merkle.KeyEncodingURL)
	keypath = keypath.AppendKey(key, merkle.KeyEncodingURL)
	keypathStr := keypath.String()

	// NOTE: the proof can specify verification rules, either only verifying the
	// top Merkle root w.r.t. all KV pairs, or verifying every layer of Merkle root
	// TODO: investigate how the verification rules are chosen when generating the
	// proof
	if err1 := prt.VerifyValue(proof, root, keypathStr, value); err1 != nil {
		if err2 := prt.VerifyAbsence(proof, root, keypathStr); err2 != nil {
			return fmt.Errorf("the Merkle proof does not pass any verification: err of VerifyValue: %w; err of VerifyAbsence: %w", err1, err2)
		}
	}

	return nil
}

func (p *ProofEpochSealed) ValidateBasic() error {
	if p.ValidatorSet == nil {
		return ErrInvalidProofEpochSealed.Wrap("ValidatorSet is nil")
	} else if len(p.ValidatorSet) == 0 {
		return ErrInvalidProofEpochSealed.Wrap("ValidatorSet is empty")
	} else if p.ProofEpochInfo == nil {
		return ErrInvalidProofEpochSealed.Wrap("ProofEpochInfo is nil")
	} else if p.ProofEpochValSet == nil {
		return ErrInvalidProofEpochSealed.Wrap("ProofEpochValSet is nil")
	}
	return nil
}

func (ih *IndexedHeader) ValidateBasic() error {
	if len(ih.ChainId) == 0 {
		return fmt.Errorf("empty ChainID")
	}
	if len(ih.Hash) == 0 {
		return fmt.Errorf("empty Hash")
	}
	if len(ih.BabylonHeaderHash) == 0 {
		return fmt.Errorf("empty BabylonHeader hash")
	}
	if len(ih.BabylonTxHash) == 0 {
		return fmt.Errorf("empty BabylonTxHash")
	}
	return nil
}

func (ih *IndexedHeader) Equal(ih2 *IndexedHeader) bool {
	if ih.ValidateBasic() != nil || ih2.ValidateBasic() != nil {
		return false
	}

	if ih.ChainId != ih2.ChainId {
		return false
	}
	if !bytes.Equal(ih.Hash, ih2.Hash) {
		return false
	}
	if ih.Height != ih2.Height {
		return false
	}
	if !bytes.Equal(ih.BabylonHeaderHash, ih2.BabylonHeaderHash) {
		return false
	}
	if ih.BabylonHeaderHeight != ih2.BabylonHeaderHeight {
		return false
	}
	if ih.BabylonEpoch != ih2.BabylonEpoch {
		return false
	}
	return bytes.Equal(ih.BabylonTxHash, ih2.BabylonTxHash)
}

func (ci *ChainInfo) Equal(ci2 *ChainInfo) bool {
	if ci.ValidateBasic() != nil || ci2.ValidateBasic() != nil {
		return false
	}

	if ci.ChainId != ci2.ChainId {
		return false
	}
	if !ci.LatestHeader.Equal(ci2.LatestHeader) {
		return false
	}
	if len(ci.LatestForks.Headers) != len(ci2.LatestForks.Headers) {
		return false
	}
	for i := 0; i < len(ci.LatestForks.Headers); i++ {
		if !ci.LatestForks.Headers[i].Equal(ci2.LatestForks.Headers[i]) {
			return false
		}
	}
	return ci.TimestampedHeadersCount == ci2.TimestampedHeadersCount
}

func (ci *ChainInfo) ValidateBasic() error {
	if len(ci.ChainId) == 0 {
		return ErrInvalidChainInfo.Wrap("ChainID is empty")
	} else if ci.LatestHeader == nil {
		return ErrInvalidChainInfo.Wrap("LatestHeader is nil")
	} else if ci.LatestForks == nil {
		return ErrInvalidChainInfo.Wrap("LatestForks is nil")
	}
	if err := ci.LatestHeader.ValidateBasic(); err != nil {
		return err
	}
	for _, forkHeader := range ci.LatestForks.Headers {
		if err := forkHeader.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func NewBTCTimestampPacketData(btcTimestamp *BTCTimestamp) *ZoneconciergePacketData {
	return &ZoneconciergePacketData{
		Packet: &ZoneconciergePacketData_BtcTimestamp{
			BtcTimestamp: btcTimestamp,
		},
	}
}
