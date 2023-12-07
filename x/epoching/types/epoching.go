package types

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewEpoch constructs a new Epoch object
// The relationship between block and epoch is as follows, assuming epoch interval of 5:
// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
// 0 |     1     |     2      |
func NewEpoch(epochNumber uint64, epochInterval uint64, firstBlockHeight uint64, lastBlockTime *time.Time) Epoch {
	return Epoch{
		EpochNumber:          epochNumber,
		CurrentEpochInterval: epochInterval,
		FirstBlockHeight:     firstBlockHeight,
		LastBlockTime:        lastBlockTime,
		// NOTE: SealerHeader will be set in the next epoch
	}
}

func (e Epoch) GetLastBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + e.CurrentEpochInterval - 1
}

func (e Epoch) GetSealerBlockHeight() uint64 {
	return e.GetLastBlockHeight() + 2
}

func (e Epoch) GetSecondBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		panic("should not be called when epoch number is zero")
	}
	return e.FirstBlockHeight + 1
}

func (e Epoch) IsLastBlock(ctx context.Context) bool {
	return e.GetLastBlockHeight() == uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
}

func (e Epoch) IsFirstBlock(ctx context.Context) bool {
	return e.FirstBlockHeight == uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
}

func (e Epoch) IsSecondBlock(ctx context.Context) bool {
	if e.EpochNumber == 0 {
		return false
	}
	return e.GetSecondBlockHeight() == uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
}

// IsFirstBlockOfNextEpoch checks whether the current block is the first block of
// the next epoch
// CONTRACT: IsFirstBlockOfNextEpoch can only be called by the epoching module
// once upon the first block of a new epoch
// other modules should use IsFirstBlock instead.
func (e Epoch) IsFirstBlockOfNextEpoch(ctx context.Context) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if e.EpochNumber == 0 {
		return sdkCtx.HeaderInfo().Height == 1
	} else {
		height := uint64(sdkCtx.HeaderInfo().Height)
		return e.FirstBlockHeight+e.CurrentEpochInterval == height
	}
}

// WithinBoundary checks whether the given height is within this epoch or not
func (e Epoch) WithinBoundary(height uint64) bool {
	if height < e.FirstBlockHeight || height > uint64(e.GetLastBlockHeight()) {
		return false
	} else {
		return true
	}
}

// ValidateBasic does sanity checks on Epoch
func (e Epoch) ValidateBasic() error {
	if e.CurrentEpochInterval < 2 {
		return ErrInvalidEpoch.Wrapf("CurrentEpochInterval (%d) < 2", e.CurrentEpochInterval)
	}
	return nil
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (qm QueuedMessage) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	msgWithType, ok := qm.UnwrapToSdkMsg().(*stakingtypes.MsgCreateValidator)
	if !ok {
		return nil
	}
	return unpacker.UnpackAny(msgWithType.Pubkey, &pubKey)
}

func (qm *QueuedMessage) UnwrapToSdkMsg() sdk.Msg {
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
		panic(errorsmod.Wrap(ErrInvalidQueuedMessageType, qm.String()))
	}
	return unwrappedMsgWithType
}
