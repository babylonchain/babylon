package types_test

import (
	"encoding/hex"
	"math/rand"
)

func genRandomByteArray(length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	rand.Read(newHeaderBytes)
	return newHeaderBytes
}

func genRandomHexStr(length uint64) string {
	randBytes := genRandomByteArray(length)
	return hex.EncodeToString(randBytes)
}

func validHex(hexStr string, length int) bool {
	if len(hexStr) != length {
		return false
	}
	if _, err := hex.DecodeString(hexStr); err != nil {
		return false
	}
	return true
}
