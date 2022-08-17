package btctxformatter

import (
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

const (
	TestTag BabylonTag = "BBT"

	MainTag BabylonTag = "BBN"

	CurrentVersion FormatVersion = 0

	firstPartNumber uint8 = 0

	secondPartNumber uint8 = 1

	HeaderLength = 4

	LastCommitHashLength = 32

	BitMapLength = 13

	AddressLength = 20

	// 8 bytes are for 64bit unsigned epoch number
	FirstHalfLength = HeaderLength + LastCommitHashLength + AddressLength + 8 + BitMapLength

	HashLength = 10

	BlsSigLength = 48

	SecondHalfLength = HeaderLength + BlsSigLength + HashLength
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

func encodeFirstTx(
	tag BabylonTag,
	version FormatVersion,
	epoch uint64,
	lastCommitHash []byte,
	bitMap []byte,
	submitterAddress []byte,
) []byte {

	var serializedBytes = []byte{}

	serializedBytes = append(serializedBytes, encodeHeader(tag, version, firstPartNumber)...)

	serializedBytes = append(serializedBytes, u64ToBEBytes(epoch)...)

	serializedBytes = append(serializedBytes, lastCommitHash...)

	serializedBytes = append(serializedBytes, bitMap...)

	serializedBytes = append(serializedBytes, submitterAddress...)

	return serializedBytes
}

func getCheckSum(firstTxBytes []byte) []byte {
	hash := sha256.Sum256(firstTxBytes)
	return hash[0:HashLength]
}

func encodeSecondTx(
	tag BabylonTag,
	version FormatVersion,
	firstTxBytes []byte,
	blsSig []byte,
) []byte {
	var serializedBytes = []byte{}

	serializedBytes = append(serializedBytes, encodeHeader(tag, version, secondPartNumber)...)

	serializedBytes = append(serializedBytes, blsSig...)

	// we are calculating checksum only from application data, without header, as header is always
	// the same.
	serializedBytes = append(serializedBytes, getCheckSum(firstTxBytes[4:])...)

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

	var firstHalf = encodeFirstTx(tag, version, epoch, lastCommitHash, bitmap, submitterAddress)

	var secondHalf = encodeSecondTx(tag, version, firstHalf, blsSig)

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
	partNumber uint8,
	data []byte,
) ([]byte, error) {

	if partNumber > 1 {
		return nil, errors.New("invalid part number")
	}

	if version > CurrentVersion {
		return nil, errors.New("invalid part number")
	}

	if partNumber == 0 && len(data) != FirstHalfLength {
		return nil, errors.New("invalid length. First part should have 77 bytes")
	}

	if partNumber == 1 && len(data) != SecondHalfLength {
		return nil, errors.New("invalid length. First part should have 62 bytes")
	}

	header := parseHeader(data)

	err := header.validateHeader(tag, version, partNumber)

	if err != nil {
		return nil, err
	}

	// At this point this is probable babylon data, strip the header and return data
	// to the caller
	dataWithoutHeader := data[4:]

	dataNoHeader := make([]byte, len(dataWithoutHeader))

	copy(dataNoHeader, dataWithoutHeader)

	return dataNoHeader, nil
}
