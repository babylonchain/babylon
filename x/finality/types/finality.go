package types

import "bytes"

func (ib *IndexedBlock) Equal(ib2 *IndexedBlock) bool {
	if !bytes.Equal(ib.Hash, ib2.Hash) {
		return false
	}
	if ib.Height != ib2.Height {
		return false
	}
	if !ib.BtcHash.Eq(ib2.BtcHash) {
		return false
	}
	if ib.BtcHeight != ib2.BtcHeight {
		return false
	}
	return true
}
