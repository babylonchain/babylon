package btctxformatter

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func randNBytes(n int) []byte {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return bytes
}

func TestEncodeMainCheckpointData(t *testing.T) {
	firstHalf, secondHalf, err := EncodeCheckpointData(
		MainTag,
		CurrentVersion,
		10,
		randNBytes(LastCommitHashLength),
		randNBytes(BitMapLength),
		randNBytes(BlsSigLength),
		randNBytes(AddressLength),
	)

	if err != nil {
		t.Errorf("Valid data should be properly encoded")
	}

	if len(firstHalf) != FirstHalfLength {
		t.Errorf("Encoded first half should have %d bytes, have %d", FirstHalfLength, len(firstHalf))
	}

	if len(secondHalf) != SecondHalfLength {
		t.Errorf("Encoded second half should have %d bytes, have %d", SecondHalfLength, len(secondHalf))
	}

	decodedFirst, err := GetCheckpointData(MainTag, CurrentVersion, 0, firstHalf)
	if err != nil {
		t.Errorf("Valid data should be properly decoded")
	}

	decodedSecond, err := GetCheckpointData(MainTag, CurrentVersion, 1, secondHalf)
	if err != nil {
		t.Errorf("Valid data should be properly decoded")
	}

	firstHalfCheckSum := sha256.Sum256(decodedFirst)

	checksumPart := firstHalfCheckSum[0:HashLength]

	checksumPartFromDec := decodedSecond[len(decodedSecond)-10:]

	if !bytes.Equal(checksumPart, checksumPartFromDec) {
		t.Errorf("Calculated checksum of first half should equal checksum attached to second half")
	}
}
