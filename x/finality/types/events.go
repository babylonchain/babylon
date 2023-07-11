package types

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
)

func NewEventSlashedBTCValidator(valBTCPK *bbn.BIP340PubKey, indexedBlock *IndexedBlock, evidence *Evidence, btcSK *btcec.PrivateKey) *EventSlashedBTCValidator {
	return &EventSlashedBTCValidator{
		ValBtcPk:       valBTCPK,
		IndexedBlock:   indexedBlock,
		Evidence:       evidence,
		ExtractedBtcSk: btcSK.Serialize(),
	}
}
