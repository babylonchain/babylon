package datagen

import (
	"encoding/hex"
	"math/rand"
)

func GenRandomByteArray(r *rand.Rand, length uint64) []byte {
	newHeaderBytes := make([]byte, length)
	r.Read(newHeaderBytes)
	return newHeaderBytes
}

func GenRandomHexStr(r *rand.Rand, length uint64) string {
	randBytes := GenRandomByteArray(r, length)
	return hex.EncodeToString(randBytes)
}

func OneInN(r *rand.Rand, n int) bool {
	return RandomInt(r, n) == 0
}

func RandomInt(r *rand.Rand, rng int) uint64 {
	return uint64(r.Intn(rng))
}

func RandomIntOtherThan(r *rand.Rand, x int, rng int) uint64 {
	if rng == 1 && x == 0 {
		panic("There is no other int")
	}
	res := RandomInt(r, rng)
	for res == uint64(x) {
		res = RandomInt(r, rng)
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
