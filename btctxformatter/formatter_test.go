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

func FuzzEncodingDecoding(f *testing.F) {
	f.Add(uint64(5), randNBytes(LastCommitHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))
	f.Add(uint64(20), randNBytes(LastCommitHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))
	f.Add(uint64(2000), randNBytes(LastCommitHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))

	f.Fuzz(func(t *testing.T, epoch uint64, lastCommitHash []byte, bitMap []byte, blsSig []byte, address []byte) {
		firstHalf, secondHalf, err := EncodeCheckpointData(
			MainTag,
			CurrentVersion,
			epoch,
			lastCommitHash,
			bitMap,
			blsSig,
			address,
		)

		if err != nil {
			// if encoding failed we cannod check anything else
			t.Skip("Encoding should be correct")
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
	})
}

// This fuzzer checks if decoder won't panic with whatever bytes we point it at
func FuzzDecodingWontPanic(f *testing.F) {
	f.Add(randNBytes(firstPartLength))
	f.Add(randNBytes(secondPartLength))

	f.Fuzz(func(t *testing.T, bytes []byte) {
		decoded, err := IsBabylonCheckpointData(MainTag, CurrentVersion, bytes)

		if err == nil {
			if decoded.index != 0 && decoded.index != 1 {
				t.Errorf("With correct decoding index should be either 0 or 1")
			}
		}
	})
}
