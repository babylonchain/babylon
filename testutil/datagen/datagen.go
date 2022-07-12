package datagen

import (
	"encoding/hex"
	"math/rand"
)

func GenRandomByteArray(length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	rand.Read(newHeaderBytes)
	return newHeaderBytes
}

func GenRandomHexStr(length uint64) string {
	randBytes := GenRandomByteArray(length)
	return hex.EncodeToString(randBytes)
}

func OneInN(n int) bool {
	return RandomInt(n) == 0
}

func RandomInt(rng int) uint64 {
	return uint64(rand.Intn(rng))
}

func RandomIntOtherThan(x int) uint64 {
	untilX := 1 + RandomInt(x)
	if RandomInt(2) == 0 {
		return uint64(x) + untilX
	}
	return uint64(x) - untilX
}

// ValidHex accepts a hex string and the length representation as a byte array
func ValidHex(hexStr string, length int) bool {
	if len(hexStr) != length*2 {
		return false
	}
	if _, err := hex.DecodeString(hexStr); err != nil {
		return false
	}
	return true
}
