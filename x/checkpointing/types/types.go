package types

import "bytes"

type BlsSigHash []byte

type RawCkptHash []byte

func (m RawCkptHash) Equals(h RawCkptHash) bool {
	if bytes.Compare(m.Bytes(), h.Bytes()) == 0 {
		return true
	}
	return false
}
