package btctxformatter

import (
	"bytes"
	cprand "crypto/rand"
	"math/rand"
	"testing"
)

func randNBytes(n int) []byte {
	bz := make([]byte, n)
	cprand.Read(bz) //nolint:errcheck // This is a test.
	return bz
}

func FuzzEncodingDecoding(f *testing.F) {
	f.Add(uint64(5), randNBytes(TagLength), randNBytes(BlockHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))
	f.Add(uint64(20), randNBytes(TagLength), randNBytes(BlockHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))
	f.Add(uint64(2000), randNBytes(TagLength), randNBytes(BlockHashLength), randNBytes(BitMapLength), randNBytes(BlsSigLength), randNBytes(AddressLength))

	f.Fuzz(func(t *testing.T, epoch uint64, tag []byte, appHash []byte, bitMap []byte, blsSig []byte, address []byte) {

		if len(tag) < TagLength {
			t.Skip("Tag should have 4 bytes")
		}

		babylonTag := BabylonTag(tag[:TagLength])

		rawBTCCkpt := &RawBtcCheckpoint{
			Epoch:            epoch,
			BlockHash:        appHash,
			BitMap:           bitMap,
			SubmitterAddress: blsSig,
			BlsSig:           address,
		}
		firstHalf, secondHalf, err := EncodeCheckpointData(
			babylonTag,
			CurrentVersion,
			rawBTCCkpt,
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

		decodedFirst, err := IsBabylonCheckpointData(babylonTag, CurrentVersion, firstHalf)

		if err != nil {
			t.Errorf("Valid data should be properly decoded")
		}

		decodedSecond, err := IsBabylonCheckpointData(babylonTag, CurrentVersion, secondHalf)

		if err != nil {
			t.Errorf("Valid data should be properly decoded")
		}

		ckptData, err := ConnectParts(CurrentVersion, decodedFirst.Data, decodedSecond.Data)
		if err != nil {
			t.Errorf("Parts should match. Error: %v", err)
		}

		ckpt, err := DecodeRawCheckpoint(CurrentVersion, ckptData)
		if err != nil {
			t.Errorf("Failed to unmarshal. Error: %v", err)
		}

		if ckpt.Epoch != epoch {
			t.Errorf("Epoch should match. Expected: %v. Got: %v", epoch, ckpt.Epoch)
		}

		if !bytes.Equal(appHash, ckpt.BlockHash) {
			t.Errorf("BlockHash should match. Expected: %v. Got: %v", appHash, ckpt.BlockHash)
		}

		if !bytes.Equal(bitMap, ckpt.BitMap) {
			t.Errorf("Bitmap should match. Expected: %v. Got: %v", bitMap, ckpt.BitMap)
		}

		if !bytes.Equal(blsSig, ckpt.BlsSig) {
			t.Errorf("BLS signature should match. Expected: %v. Got: %v", blsSig, ckpt.BlsSig)
		}
	})
}

// This fuzzer checks if decoder won't panic with whatever bytes we point it at
func FuzzDecodingWontPanic(f *testing.F) {
	f.Add(randNBytes(firstPartLength), uint8(rand.Intn(99)))
	f.Add(randNBytes(secondPartLength), uint8(rand.Intn(99)))

	f.Fuzz(func(t *testing.T, bytes []byte, tagIdx uint8) {
		tag := []byte{0, 1, 2, 3}
		decoded, err := IsBabylonCheckpointData(tag, CurrentVersion, bytes)

		if err == nil {
			if decoded.Index != 0 && decoded.Index != 1 {
				t.Errorf("With correct decoding index should be either 0 or 1")
			}
		}
	})
}
