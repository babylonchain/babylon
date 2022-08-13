package simulation_test

import (
	"fmt"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"testing"

	"github.com/stretchr/testify/require"

	bbnapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/simulation"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// nolint:deadcode,unused,varcheck
var (
	delPk1      = ed25519.GenPrivKey().PubKey()
	delAddr1    = sdk.AccAddress(delPk1.Address())
	valAddr1    = sdk.ValAddress(delPk1.Address())
	consAddr1   = sdk.ConsAddress(delPk1.Address().Bytes())
	oneBytes, _ = sdk.NewInt(1).Marshal()
)

func TestDecodeStore(t *testing.T) {
	cdc := bbnapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	epochNumber := uint64(123)
	queuedMsg := types.QueuedMessage{
		TxId:  sdk.Uint64ToBigEndian(333),
		MsgId: sdk.Uint64ToBigEndian(444),
		Msg:   &types.QueuedMessage_MsgDelegate{MsgDelegate: &stakingtypes.MsgDelegate{}},
	}
	valSet := types.ValidatorSet{
		types.Validator{
			Addr:  valAddr1,
			Power: 1,
		},
	}

	marshaledQueueMsg, err := cdc.MarshalInterface(&queuedMsg)
	require.NoError(t, err)
	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.EpochNumberKey, Value: sdk.Uint64ToBigEndian(epochNumber)},
			{Key: types.QueuedMsgKey, Value: marshaledQueueMsg},
			{Key: types.ValidatorSetKey, Value: valSet.MustMarshal()},
			{Key: types.SlashedValidatorSetKey, Value: valSet.MustMarshal()},
			{Key: types.VotingPowerKey, Value: oneBytes},
			{Key: types.SlashedVotingPowerKey, Value: oneBytes},
			{Key: []byte{0x99}, Value: []byte{0x99}}, // This test should panic
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"EpochNumber", fmt.Sprintf("%v\n%v", epochNumber, epochNumber)},
		{"QueuedMsg", fmt.Sprintf("%v\n%v", queuedMsg.MsgId, queuedMsg.MsgId)},
		{"ValidatorSet", fmt.Sprintf("%v\n%v", valSet, valSet)},
		{"SlashedValidatorSet", fmt.Sprintf("%v\n%v", valSet, valSet)},
		{"VotingPower", fmt.Sprintf("%v\n%v", sdk.NewInt(1), sdk.NewInt(1))},
		{"SlashedVotingPower", fmt.Sprintf("%v\n%v", sdk.NewInt(1), sdk.NewInt(1))},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
