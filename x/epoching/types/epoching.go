package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func NewEpoch(epochNumber uint64, epochInterval uint64) Epoch {
	return Epoch{
		EpochNumber:          epochNumber,
		CurrentEpochInterval: epochInterval,
		FirstBlockHeight:     firstBlockHeight(epochNumber, epochInterval),
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
	switch msg := msg.(type) {
	case *MsgWrappedDelegate:
		if msgBytes, err = msg.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgDelegate{
			MsgDelegate: msg.Msg,
		}
	case *MsgWrappedBeginRedelegate:
		if msgBytes, err = msg.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgBeginRedelegate{
			MsgBeginRedelegate: msg.Msg,
		}
	case *MsgWrappedUndelegate:
		if msgBytes, err = msg.Msg.Marshal(); err != nil {
			return QueuedMessage{}, err
		}
		qmsg = &QueuedMessage_MsgUndelegate{
			MsgUndelegate: msg.Msg,
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
