package types_test

import (
	"github.com/babylonchain/babylon/types"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
	"time"
)

type btcutilsTestSuite struct {
	suite.Suite
	mainnetHeader, testnetHeader, mainnetHeaderInvalidTs *wire.BlockHeader
	mainnetPowLimit, testnetPowLimit                     *big.Int
}

func TestBtcutilsTestSuite(t *testing.T) {
	suite.Run(t, new(btcutilsTestSuite))
}

func (s *btcutilsTestSuite) SetupSuite() {
	s.T().Parallel()
	mainnetHeaderHex := "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22"
	testnetHeaderHex := "0000202015f76d6c7147a1cd3cca4ec31b75ac0218199e863ebd040c6000000000000000d321bbe2d2e323781b4ab89abc24d37c6ad70d92a5169f9775b650447548711162957a626fa4001a90376af4"
	mainnetHeaderBytes, _ := types.NewBTCHeaderBytesFromHex(mainnetHeaderHex)
	testnetHeaderBytes, _ := types.NewBTCHeaderBytesFromHex(testnetHeaderHex)

	s.mainnetHeader = mainnetHeaderBytes.ToBlockHeader()
	s.testnetHeader = testnetHeaderBytes.ToBlockHeader()

	mainnetHeader := *s.mainnetHeader
	s.mainnetHeaderInvalidTs = &mainnetHeader
	s.mainnetHeaderInvalidTs.Timestamp = time.Now()

	s.mainnetPowLimit = btcchaincfg.MainNetParams.PowLimit
	s.testnetPowLimit = btcchaincfg.TestNet3Params.PowLimit
}

func (s *btcutilsTestSuite) TestValidateBTCHeader() {
	data := []struct {
		name     string
		header   *wire.BlockHeader
		powLimit *big.Int
		hasErr   bool
	}{
		{"valid mainnet", s.mainnetHeader, s.mainnetPowLimit, false},
		{"valid testnet", s.testnetHeader, s.testnetPowLimit, false},
		{"mainnet invalid limit", s.mainnetHeader, big.NewInt(0), true},
		{"testnet invalid limit", s.testnetHeader, big.NewInt(0), true},
		{"mainnet invalid timestamp", s.mainnetHeaderInvalidTs, s.mainnetPowLimit, true},
	}

	for _, d := range data {
		err := types.ValidateBTCHeader(d.header, d.powLimit)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
		}
	}
}
