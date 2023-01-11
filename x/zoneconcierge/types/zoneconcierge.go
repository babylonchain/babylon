package types

import (
	"bytes"
	"fmt"
)

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
	} else if len(ih.Hash) == 0 {
		return fmt.Errorf("empty Hash")
	} else if ih.BabylonHeader == nil {
		return fmt.Errorf("nil BabylonHeader")
	} else if len(ih.BabylonTxHash) == 0 {
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
	} else if !bytes.Equal(ih.Hash, ih2.Hash) {
		return false
	} else if ih.Height != ih2.Height {
		return false
	} else if !bytes.Equal(ih.BabylonHeader.LastCommitHash, ih2.BabylonHeader.LastCommitHash) {
		return false
	} else if ih.BabylonEpoch != ih2.BabylonEpoch {
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
	}
	return nil
}
