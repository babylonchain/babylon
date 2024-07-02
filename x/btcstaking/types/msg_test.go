package types_test

import (
	"fmt"
	"math/rand"
	"testing"

	"cosmossdk.io/errors"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbntypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stktypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateFinalityProviderValidateBasic(t *testing.T) {
	r := rand.New(rand.NewSource(10))
	randBigMoniker := datagen.GenRandomHexStr(r, 100)

	bigBtcPK := datagen.GenRandomByteArray(r, 100)

	fp, err := datagen.GenRandomFinalityProvider(r)
	require.NoError(t, err)

	invalidAddr := "bbnbadaddr"

	tcs := []struct {
		title  string
		msg    *types.MsgCreateFinalityProvider
		expErr error
	}{
		{
			"valid: msg create fp",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop:         fp.Pop,
			},
			nil,
		},
		{
			"invalid: empty commission",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  nil,
				BtcPk:       fp.BtcPk,
				Pop:         fp.Pop,
			},
			fmt.Errorf("empty commission"),
		},
		{
			"invalid: empty description",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: nil,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop:         fp.Pop,
			},
			fmt.Errorf("empty description"),
		},
		{
			"invalid: empty moniker",
			&types.MsgCreateFinalityProvider{
				Addr: fp.Addr,
				Description: &stktypes.Description{
					Moniker:         "",
					Identity:        fp.Description.Identity,
					Website:         fp.Description.Website,
					SecurityContact: fp.Description.SecurityContact,
					Details:         fp.Description.Details,
				},
				Commission: fp.Commission,
				BtcPk:      fp.BtcPk,
				Pop:        fp.Pop,
			},
			fmt.Errorf("empty moniker"),
		},
		{
			"invalid: big moniker",
			&types.MsgCreateFinalityProvider{
				Addr: fp.Addr,
				Description: &stktypes.Description{
					Moniker:         randBigMoniker,
					Identity:        fp.Description.Identity,
					Website:         fp.Description.Website,
					SecurityContact: fp.Description.SecurityContact,
					Details:         fp.Description.Details,
				},
				Commission: fp.Commission,
				BtcPk:      fp.BtcPk,
				Pop:        fp.Pop,
			},
			errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(randBigMoniker), stktypes.MaxMonikerLength),
		},
		{
			"invalid: empty BTC pk",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       nil,
				Pop:         fp.Pop,
			},
			fmt.Errorf("empty BTC public key"),
		},
		{
			"invalid: invalid BTC pk",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       (*bbntypes.BIP340PubKey)(&bigBtcPK),
				Pop:         fp.Pop,
			},
			fmt.Errorf("invalid BTC public key: %v", fmt.Errorf("bad pubkey byte string size (want %v, have %v)", 32, len(bigBtcPK))),
		},
		{
			"invalid: empty PoP",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop:         nil,
			},
			fmt.Errorf("empty proof of possession"),
		},
		{
			"invalid: empty PoP",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop:         nil,
			},
			fmt.Errorf("empty proof of possession"),
		},
		{
			"invalid: bad addr",
			&types.MsgCreateFinalityProvider{
				Addr:        invalidAddr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop:         fp.Pop,
			},
			fmt.Errorf("invalid FP addr: %s - %v", invalidAddr, fmt.Errorf("decoding bech32 failed: invalid separator index -1")),
		},
		{
			"invalid: bad PoP empty sig",
			&types.MsgCreateFinalityProvider{
				Addr:        fp.Addr,
				Description: fp.Description,
				Commission:  fp.Commission,
				BtcPk:       fp.BtcPk,
				Pop: &types.ProofOfPossessionBTC{
					BtcSig: nil,
				},
			},
			fmt.Errorf("empty BTC signature"),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			actErr := tc.msg.ValidateBasic()
			if tc.expErr != nil {
				require.EqualError(t, actErr, tc.expErr.Error())
				return
			}
			require.NoError(t, actErr)
		})
	}
}
