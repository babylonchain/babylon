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

func RandomIntOtherThan(x int, rng int) uint64 {
	if rng == 1 && x == 0 {
		panic("There is no other int")
	}
	res := RandomInt(rng)
	for res == uint64(x) {
		res = RandomInt(rng)
	}
	return res
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
