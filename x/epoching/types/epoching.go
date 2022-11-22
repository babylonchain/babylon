package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func NewEpoch(epochNumber uint64, epochInterval uint64, lastBlockHeader *tmproto.Header) Epoch {
	return Epoch{
		EpochNumber:          epochNumber,
		CurrentEpochInterval: epochInterval,
		FirstBlockHeight:     firstBlockHeight(epochNumber, epochInterval),
		LastBlockHeader:      lastBlockHeader,
	}
}

// firstBlockHeight returns the height of the first block of a given epoch and epoch interval
// TODO (non-urgent): add support to variable epoch interval
func firstBlockHeight(epochNumber uint64, epochInterval uint64) uint64 {
	// example: in epoch 2, epoch interval is 5 blocks, FirstBlockHeight will be (2-1)*5+1 = 6
	// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
	// 0 |     1     |     2      |
	if epochNumber == 0 {
		return 0
	} else {
		return (epochNumber-1)*epochInterval + 1
	}
}

func (e Epoch) GetLastBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + e.CurrentEpochInterval - 1
}

func (e Epoch) GetSecondBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		panic("should not be called when epoch number is zero")
	}
	return e.FirstBlockHeight + 1
}

func (e Epoch) IsLastBlock(ctx sdk.Context) bool {
	return e.GetLastBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlock(ctx sdk.Context) bool {
	return e.FirstBlockHeight == uint64(ctx.BlockHeight())
}

func (e Epoch) IsSecondBlock(ctx sdk.Context) bool {
	if e.EpochNumber == 0 {
		return false
	}
	return e.GetSecondBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlockOfNextEpoch(ctx sdk.Context) bool {
	if e.EpochNumber == 0 {
		return ctx.BlockHeight() == 1
	} else {
		height := uint64(ctx.BlockHeight())
		return e.FirstBlockHeight+e.CurrentEpochInterval == height
	}
}

// NewQueuedMessage creates a new QueuedMessage from a wrapped msg
// i.e., wrapped -> unwrapped -> QueuedMessage
func NewQueuedMessage(blockHeight uint64, blockTime time.Time, txid []byte, msg sdk.Msg) (QueuedMessage, error) {
	// marshal the actual msg (MsgDelegate, MsgBeginRedelegate, MsgUndelegate, ...) inside isQueuedMessage_Msg
	// TODO (non-urgent): after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
	var qmsg isQueuedMessage_Msg
	var msgBytes []byte
	var err error
	switch msgWithType := msg.(type) {
	case *MsgWrappedDelegate:
		if msgBytes, err = msgWithType.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgDelegate{
			MsgDelegate: msgWithType.Msg,
		}
	case *MsgWrappedBeginRedelegate:
		if msgBytes, err = msgWithType.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgBeginRedelegate{
			MsgBeginRedelegate: msgWithType.Msg,
		}
	case *MsgWrappedUndelegate:
		if msgBytes, err = msgWithType.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgUndelegate{
			MsgUndelegate: msgWithType.Msg,
		}
	case *stakingtypes.MsgCreateValidator:
		if msgBytes, err = msgWithType.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgCreateValidator{
			MsgCreateValidator: msgWithType,
		}
	default:
		return QueuedMessage{}, ErrUnwrappedMsgType
	}

	queuedMsg := QueuedMessage{
		TxId:        txid,
		MsgId:       tmhash.Sum(msgBytes),
		BlockHeight: blockHeight,
		BlockTime:   &blockTime,
		Msg:         qmsg,
	}
	return queuedMsg, nil
}

func (qm QueuedMessage) GetSigners() []sdk.AccAddress {
	return qm.WithType().GetSigners()
}

func (qm QueuedMessage) ValidateBasic() error {
	return qm.WithType().ValidateBasic()

}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (qm QueuedMessage) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	msgWithType, ok := qm.WithType().(*stakingtypes.MsgCreateValidator)
	if !ok {
		return nil
	}
	return unpacker.UnpackAny(msgWithType.Pubkey, &pubKey)
}

func (qm *QueuedMessage) WithType() sdk.Msg {
	var unwrappedMsgWithType sdk.Msg
	// TODO (non-urgent): after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
	switch unwrappedMsg := qm.Msg.(type) {
	case *QueuedMessage_MsgCreateValidator:
		unwrappedMsgWithType = unwrappedMsg.MsgCreateValidator
	case *QueuedMessage_MsgDelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgDelegate
	case *QueuedMessage_MsgUndelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgUndelegate
	case *QueuedMessage_MsgBeginRedelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgBeginRedelegate
	default:
		panic(sdkerrors.Wrap(ErrInvalidQueuedMessageType, qm.String()))
	}
	return unwrappedMsgWithType
}
