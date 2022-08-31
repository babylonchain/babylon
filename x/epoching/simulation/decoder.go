package simulation

import (
	"bytes"
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding epoching type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.EpochNumberKey),
			bytes.Equal(kvA.Key[:1], types.QueueLengthKey):
			return fmt.Sprintf("%v\n%v", sdk.BigEndianToUint64(kvA.Value), sdk.BigEndianToUint64(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.MsgQueueKey):
			var qmA, qmB types.QueuedMessage
			cdc.MustUnmarshal(kvA.Value, &qmA)
			cdc.MustUnmarshal(kvB.Value, &qmB)
			return fmt.Sprintf("%v\n%v", qmA, qmB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorSetKey),
			bytes.Equal(kvA.Key[:1], types.SlashedValidatorSetKey):
			valSetA, err := types.NewValidatorSetFromBytes(kvA.Value)
			if err != nil {
				panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
			}
			valSetB, err := types.NewValidatorSetFromBytes(kvB.Value)
			if err != nil {
				panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
			}
			return fmt.Sprintf("%v\n%v", valSetA, valSetB)

		case bytes.Equal(kvA.Key[:1], types.VotingPowerKey),
			bytes.Equal(kvA.Key[:1], types.SlashedVotingPowerKey):
			var powerA, powerB sdk.Int
			if err := powerA.Unmarshal(kvA.Value); err != nil {
				panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
			}
			if err := powerB.Unmarshal(kvA.Value); err != nil {
				panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
			}
			return fmt.Sprintf("%v\n%v", powerA, powerB)

		default:
			panic(fmt.Sprintf("invalid epoching key prefix %X", kvA.Key[:1]))
		}
	}
}
