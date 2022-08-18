package btctxformatter

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

type BabylonTag string

type FormatVersion uint8

type formatHeader struct {
	tag     BabylonTag
	version FormatVersion
	part    uint8
}

type BabylonData struct {
	data  []byte
	index uint8
}

const (
	TestTag BabylonTag = "BBT"

	MainTag BabylonTag = "BBN"

	CurrentVersion FormatVersion = 0

	firstPartIndex uint8 = 0

	secondPartIndex uint8 = 1

	headerLength = 4

	LastCommitHashLength = 32

	BitMapLength = 13

	AddressLength = 20

	// Each checkpoint is composed of two parts
	NumberOfParts = 2

	hashLength = 10

	BlsSigLength = 48

	// 8 bytes are for 64bit unsigned epoch number
	firstPartLength = headerLength + LastCommitHashLength + AddressLength + 8 + BitMapLength

	secondPartLength = headerLength + BlsSigLength + hashLength

	ApplicationDataLength = firstPartLength + secondPartLength - headerLength - headerLength - hashLength
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

func u64ToBEBytes(u uint64) []byte {
	bytes := make([]byte, 8)

	binary.BigEndian.PutUint64(bytes, u)

	return bytes
}

func encodeFirstOpRetrun(
	tag BabylonTag,
	version FormatVersion,
	epoch uint64,
	lastCommitHash []byte,
	bitMap []byte,
	submitterAddress []byte,
) []byte {

	var serializedBytes = []byte{}

	serializedBytes = append(serializedBytes, encodeHeader(tag, version, firstPartIndex)...)

	serializedBytes = append(serializedBytes, u64ToBEBytes(epoch)...)

	serializedBytes = append(serializedBytes, lastCommitHash...)

	serializedBytes = append(serializedBytes, bitMap...)

	serializedBytes = append(serializedBytes, submitterAddress...)

	return serializedBytes
}

func getCheckSum(firstTxBytes []byte) []byte {
	hash := sha256.Sum256(firstTxBytes)
	return hash[0:hashLength]
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
	epoch uint64,
	lastCommitHash []byte,
	bitmap []byte,
	blsSig []byte,
	submitterAddress []byte,
) ([]byte, []byte, error) {

	if tag != MainTag && tag != TestTag {
		return nil, nil, errors.New("not allowed Tag value")
	}

	if version > CurrentVersion {
		return nil, nil, errors.New("invalid format version")
	}

	if len(lastCommitHash) != LastCommitHashLength {
		return nil, nil, errors.New("lastCommitHash should have 32 bytes")
	}

	if len(bitmap) != BitMapLength {
		return nil, nil, errors.New("bitmap should have 13 bytes")
	}

	if len(blsSig) != BlsSigLength {
		return nil, nil, errors.New("BlsSig should have 48 bytes")
	}

	if len(blsSig) != BlsSigLength {
		return nil, nil, errors.New("BlsSig should have 48 bytes")
	}

	var firstHalf = encodeFirstOpRetrun(tag, version, epoch, lastCommitHash, bitmap, submitterAddress)

	var secondHalf = encodeSecondOpReturn(tag, version, firstHalf, blsSig)

	return firstHalf, secondHalf, nil
}

func MustEncodeCheckpointData(
	tag BabylonTag,
	version FormatVersion,
	epoch uint64,
	lastCommitHash []byte,
	bitmap []byte,
	blsSig []byte,
	submitterAddress []byte,
) ([]byte, []byte) {
	f, s, err := EncodeCheckpointData(tag, version, epoch, lastCommitHash, bitmap, blsSig, submitterAddress)
	if err != nil {
		panic(err)
	}

	return f, s
}

func parseHeader(
	data []byte,
) *formatHeader {
	tagBytes := data[0:3]

	verHalf := data[3]

	header := formatHeader{
		tag:     BabylonTag(tagBytes),
		version: FormatVersion((verHalf & 0xf)),
		part:    verHalf >> 4,
	}

	return &header
}

func (header *formatHeader) validateHeader(
	expectedTag BabylonTag,
	supportedVersion FormatVersion,
	expectedPart uint8,
) error {
	if header.tag != expectedTag {
		return errors.New("data does not have expected tag")
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
			bd := BabylonData{data: data, index: idx}
			return &bd, nil
		}

		idx++
	}

	return nil, errors.New("not valid babylon data")
}

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

	hashStartIdx := len(s) - hashLength

	expectedHash := s[hashStartIdx:]

	if !bytes.Equal(firstHash[:hashLength], expectedHash) {
		return nil, errors.New("parts do not connect")
	}

	var dst []byte
	// TODO this is not supper efficient
	dst = append(dst, f...)
	dst = append(dst, s[:hashStartIdx]...)

	return dst, nil
}
