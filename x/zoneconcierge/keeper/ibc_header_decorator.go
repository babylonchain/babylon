package keeper

import (
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

type HeaderInfo struct {
	ChainId string
	Hash    []byte
	Height  uint64
}

func (d IBCHeaderDecorator) getHeaderAndClientState(ctx sdk.Context, m sdk.Msg) (*HeaderInfo, *ibctmtypes.ClientState) {
	// ensure the message is MsgUpdateClient
	msgUpdateClient, ok := m.(*clienttypes.MsgUpdateClient)
	if !ok {
		return nil, nil
	}
	// unpack ClientMsg inside MsgUpdateClient
	clientMsg, err := clienttypes.UnpackClientMessage(msgUpdateClient.ClientMessage)
	if err != nil {
		return nil, nil
	}
	// ensure the ClientMsg is a Tendermint header
	ibctmHeader, ok := clientMsg.(*ibctmtypes.Header)
	if !ok {
		return nil, nil
	}

	// all good, we get the headerInfo
	headerInfo := &HeaderInfo{
		ChainId: ibctmHeader.Header.ChainID,
		Hash:    ibctmHeader.Header.LastCommitHash,
		Height:  uint64(ibctmHeader.Header.Height),
	}

	// ensure the corresponding clientState exists
	clientState, exist := d.k.clientKeeper.GetClientState(ctx, msgUpdateClient.ClientId)
	if !exist {
		return nil, nil
	}
	// ensure the clientState is a Tendermint clientState
	tmClientState, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return nil, nil
	}

	return headerInfo, tmClientState
}

type IBCHeaderDecorator struct {
	k Keeper
}

// NewIBCHeaderDecorator creates a new IBCHeaderDecorator
func NewIBCHeaderDecorator(k Keeper) *IBCHeaderDecorator {
	return &IBCHeaderDecorator{
		k: k,
	}
}

func (d IBCHeaderDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		// try to extract the headerInfo and the client's status
		headerInfo, clientState := d.getHeaderAndClientState(ctx, msg)
		if headerInfo == nil {
			continue
		}

		txHash := tmhash.Sum(ctx.TxBytes())

		// FrozenHeight is non-zero -> client is frozen -> this is a fork header
		isOnFork := !clientState.FrozenHeight.IsZero()
		d.k.Hooks().AfterHeaderWithValidCommit(ctx, txHash, headerInfo, isOnFork)

		// unfreeze client (by setting FrozenHeight to zero again)
		if isOnFork {
			clientState.FrozenHeight = clienttypes.ZeroHeight()
		}
	}

	return next(ctx, tx, simulate, success)
}
