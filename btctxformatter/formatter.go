package btctxformatter

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

type BabylonTag []byte

type FormatVersion uint8

type formatHeader struct {
	tag     BabylonTag
	version FormatVersion
	part    uint8
}

type BabylonData struct {
	Data  []byte
	Index uint8
}

type RawBtcCheckpoint struct {
	Epoch            uint64
	BlockHash        []byte
	BitMap           []byte
	SubmitterAddress []byte
	BlsSig           []byte
}

const (
	TagLength = 4

	CurrentVersion FormatVersion = 0

	firstPartIndex uint8 = 0

	secondPartIndex uint8 = 1

	// 4bytes tag + 4 bits version + 4 bits part index
	headerLength = TagLength + 1

	BlockHashLength = 32

	// BitMapLength is the number of bytes in a bitmap
	// It is the minimal number needed for supporting 100
	// validators in BTC timestamping, since 13*8 = 104
	BitMapLength = 13

	AddressLength = 20

	// Each checkpoint is composed of two parts
	NumberOfParts = 2

	// First 10 bytes of sha256 of first part are appended to second part to ease up
	// pairing of parts
	firstPartHashLength = 10

	BlsSigLength = 48

	// 8 bytes are for 64bit unsigned epoch number
	EpochLength = 8

	firstPartLength = headerLength + BlockHashLength + AddressLength + EpochLength + BitMapLength

	secondPartLength = headerLength + BlsSigLength + firstPartHashLength

	RawBTCCheckpointLength = EpochLength + BlockHashLength + BitMapLength + BlsSigLength + AddressLength
)

func getVerHalf(version FormatVersion, halfNumber uint8) uint8 {
	var verHalf = uint8(0)
	// set first 4bits as version
	verHalf = (verHalf & 0xf0) | (uint8(version) & 0xf)
	// set last 4bits as half number
	verHalf = (verHalf & 0xf) | (halfNumber << uint8(4))

	return verHalf
}

func encodeHeader(tag BabylonTag, version FormatVersion, halfNumber uint8) []byte {
	var data = []byte(tag)
	data = append(data, getVerHalf(version, halfNumber))
	return data
}

func U64ToBEBytes(u uint64) []byte {
	u64bytes := make([]byte, 8)

	binary.BigEndian.PutUint64(u64bytes, u)

	return u64bytes
}

func encodeFirstOpRetrun(
	tag BabylonTag,
	version FormatVersion,
	epoch uint64,
	appHash []byte,
	bitMap []byte,
	submitterAddress []byte,
) []byte {

	var serializedBytes = []byte{}

	serializedBytes = append(serializedBytes, encodeHeader(tag, version, firstPartIndex)...)

	serializedBytes = append(serializedBytes, U64ToBEBytes(epoch)...)

	serializedBytes = append(serializedBytes, appHash...)

	serializedBytes = append(serializedBytes, bitMap...)

	serializedBytes = append(serializedBytes, submitterAddress...)

	return serializedBytes
}

func getCheckSum(firstTxBytes []byte) []byte {
	hash := sha256.Sum256(firstTxBytes)
	return hash[0:firstPartHashLength]
}

func encodeSecondOpReturn(
	tag BabylonTag,
	version FormatVersion,
	firstOpReturnBytes []byte,
	blsSig []byte,
) []byte {
	var serializedBytes = []byte{}

	serializedBytes = append(serializedBytes, encodeHeader(tag, version, secondPartIndex)...)

	serializedBytes = append(serializedBytes, blsSig...)

	// we are calculating checksum only from application data, without header, as header is always
	// the same.
	serializedBytes = append(serializedBytes, getCheckSum(firstOpReturnBytes[headerLength:])...)

	return serializedBytes
}

func EncodeCheckpointData(
	tag BabylonTag,
	version FormatVersion,
	rawBTCCheckpoint *RawBtcCheckpoint,
) ([]byte, []byte, error) {

	if len(tag) != TagLength {
		return nil, nil, errors.New("tag should have 4 bytes")
	}

	if version > CurrentVersion {
		return nil, nil, errors.New("invalid format version")
	}

	if len(rawBTCCheckpoint.BlockHash) != BlockHashLength {
		return nil, nil, errors.New("appHash should have 32 bytes")
	}

	if len(rawBTCCheckpoint.BitMap) != BitMapLength {
		return nil, nil, errors.New("bitmap should have 13 bytes")
	}

	if len(rawBTCCheckpoint.BlsSig) != BlsSigLength {
		return nil, nil, errors.New("BlsSig should have 48 bytes")
	}

	if len(rawBTCCheckpoint.SubmitterAddress) != AddressLength {
		return nil, nil, errors.New("address should have 20 bytes")
	}

	var firstHalf = encodeFirstOpRetrun(
		tag,
		version,
		rawBTCCheckpoint.Epoch,
		rawBTCCheckpoint.BlockHash,
		rawBTCCheckpoint.BitMap,
		rawBTCCheckpoint.SubmitterAddress,
	)

	var secondHalf = encodeSecondOpReturn(
		tag,
		version,
		firstHalf,
		rawBTCCheckpoint.BlsSig,
	)

	return firstHalf, secondHalf, nil
}

func MustEncodeCheckpointData(
	tag BabylonTag,
	version FormatVersion,
	rawBTCCheckpoint *RawBtcCheckpoint,
) ([]byte, []byte) {
	f, s, err := EncodeCheckpointData(tag, version, rawBTCCheckpoint)
	if err != nil {
		panic(err)
	}

	return f, s
}

func parseHeader(
	data []byte,
) *formatHeader {
	tagBytes := data[:TagLength]

	verHalf := data[TagLength]

	header := formatHeader{
		tag:     BabylonTag(tagBytes),
		version: FormatVersion(verHalf & 0xf),
		part:    verHalf >> 4,
	}

	return &header
}

func (header *formatHeader) validateHeader(
	expectedTag BabylonTag,
	_ FormatVersion,
	expectedPart uint8,
) error {
	if !bytes.Equal(header.tag, expectedTag) {
		return fmt.Errorf("data does not have expected tag, expected tag: %v, got tag: %v", expectedTag, header.tag)
	}

	if header.version > CurrentVersion {
		return errors.New("header have invalid version")
	}

	if header.part != expectedPart {
		return errors.New("header have invalid part number")
	}

	return nil
}

func GetCheckpointData(
	tag BabylonTag,
	version FormatVersion,
	partIndex uint8,
	data []byte,
) ([]byte, error) {

	if partIndex > 1 {
		return nil, errors.New("invalid part index")
	}

	if version > CurrentVersion {
		return nil, errors.New("not supported version")
	}

	if partIndex == 0 && len(data) != firstPartLength {
		return nil, errors.New("invalid length. First part should have 77 bytes")
	}

	if partIndex == 1 && len(data) != secondPartLength {
		return nil, errors.New("invalid length. First part should have 62 bytes")
	}

	header := parseHeader(data)

	err := header.validateHeader(tag, version, partIndex)

	if err != nil {
		return nil, err
	}

	// At this point this is probable babylon data, strip the header and return data
	// to the caller
	dataWithoutHeader := data[headerLength:]

	dataNoHeader := make([]byte, len(dataWithoutHeader))

	copy(dataNoHeader, dataWithoutHeader)

	return dataNoHeader, nil
}

// IsBabylonCheckpointData Checks if given bytearray is potential babylon data,
// if it is then returns index of data along side with data itself
func IsBabylonCheckpointData(
	tag BabylonTag,
	version FormatVersion,
	data []byte,
) (*BabylonData, error) {

	var idx uint8 = 0

	for idx < NumberOfParts {
		data, err := GetCheckpointData(tag, version, idx, data)

		if err == nil {
			bd := BabylonData{Data: data, Index: idx}
			return &bd, nil
		}

		idx++
	}

	return nil, errors.New("not valid babylon data")
}

// DecodeRawCheckpoint extracts epoch, appHash, bitmap, and blsSig from a
// flat byte array and compose them into a RawCheckpoint struct
func DecodeRawCheckpoint(version FormatVersion, btcCkptBytes []byte) (*RawBtcCheckpoint, error) {
	if version > CurrentVersion {
		return nil, errors.New("not supported version")
	}

	if len(btcCkptBytes) != RawBTCCheckpointLength {
		return nil, errors.New("invalid raw checkpoint data length")
	}

	var b bytes.Buffer
	b.Write(btcCkptBytes)
	epochBytes := b.Next(EpochLength)
	appHashBytes := b.Next(BlockHashLength)
	bitmapBytes := b.Next(BitMapLength)
	addressBytes := b.Next(AddressLength)
	blsSigBytes := b.Next(BlsSigLength)

	rawCheckpoint := &RawBtcCheckpoint{
		Epoch:            binary.BigEndian.Uint64(epochBytes),
		BlockHash:        appHashBytes,
		BitMap:           bitmapBytes,
		SubmitterAddress: addressBytes,
		BlsSig:           blsSigBytes,
	}

	return rawCheckpoint, nil
}

// ConnectParts composes raw checkpoint data by connecting two parts
// of checkpoint data and stripping off data that is not relevant to a raw checkpoint
func ConnectParts(version FormatVersion, f []byte, s []byte) ([]byte, error) {
	if version > CurrentVersion {
		return nil, errors.New("not supported version")
	}

	if len(f) != firstPartLength-headerLength {
		return nil, errors.New("not valid first part")
	}

	if len(s) != secondPartLength-headerLength {
		return nil, errors.New("not valid second part")
	}

	firstHash := sha256.Sum256(f)

	hashStartIdx := len(s) - firstPartHashLength

	expectedHash := s[hashStartIdx:]

	if !bytes.Equal(firstHash[:firstPartHashLength], expectedHash) {
		return nil, errors.New("parts do not connect")
	}

	var dst []byte
	// TODO this is not supper efficient
	dst = append(dst, f...)
	dst = append(dst, s[:hashStartIdx]...)

	return dst, nil
}
