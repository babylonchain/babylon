package btctxformatter

import (
	"crypto/rand"
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

	if len(firstHalf) != firstPartLength {
		t.Errorf("Encoded first half should have %d bytes, have %d", firstPartLength, len(firstHalf))
	}

	if len(secondHalf) != secondPartLength {
		t.Errorf("Encoded second half should have %d bytes, have %d", secondPartLength, len(secondHalf))
	}

	decodedFirst, err := IsBabylonCheckpointData(MainTag, CurrentVersion, firstHalf)

	if err != nil {
		t.Errorf("Valid data should be properly decoded")
	}

	decodedSecond, err := IsBabylonCheckpointData(MainTag, CurrentVersion, secondHalf)

	if err != nil {
		t.Errorf("Valid data should be properly decoded")
	}

	data, err := ConnectParts(CurrentVersion, decodedFirst.data, decodedSecond.data)

	if err != nil {
		t.Errorf("Parts should match. Error: %v", err)
	}

	if len(data) != ApplicationDataLength {
		t.Errorf("Not expected application level data length. Have: %d, want: %d", len(data), ApplicationDataLength)
	}
}
