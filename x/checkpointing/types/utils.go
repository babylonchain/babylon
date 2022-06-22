package types

import (
	"crypto/sha256"
	"fmt"
)

func (m BlsSig) Hash() BlsSigHash {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", m)))

	return h.Sum(nil)
}

func (m RawCheckpoint) Hash() RawCkptHash {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", m)))

	return h.Sum(nil)
}
