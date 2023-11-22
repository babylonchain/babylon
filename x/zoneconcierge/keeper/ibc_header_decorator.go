package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

func (d IBCHeaderDecorator) getHeaderAndClientState(ctx sdk.Context, m sdk.Msg) (*types.HeaderInfo, *ibctmtypes.ClientState) {
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
	// ensure the ClientMsg is a Comet header
	ibctmHeader, ok := clientMsg.(*ibctmtypes.Header)
	if !ok {
		return nil, nil
	}

	// all good, we get the headerInfo
	headerInfo := &types.HeaderInfo{
		ClientId: msgUpdateClient.ClientId,
		ChainId:  ibctmHeader.Header.ChainID,
		Hash:     ibctmHeader.Header.AppHash,
		Height:   uint64(ibctmHeader.Header.Height),
		Time:     ibctmHeader.Header.Time,
	}

	// ensure the corresponding clientState exists
	clientState, exist := d.k.clientKeeper.GetClientState(ctx, msgUpdateClient.ClientId)
	if !exist {
		return nil, nil
	}
	// ensure the clientState is a Comet clientState
	cmtClientState, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return nil, nil
	}

	return headerInfo, cmtClientState
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
	// ignore unsuccessful tx
	// NOTE: tx with a misbehaving header will still succeed, but will make the client to be frozen
	if !success {
		return next(ctx, tx, simulate, success)
	}

	// calculate tx hash
	txHash := tmhash.Sum(ctx.TxBytes())

	for _, msg := range tx.GetMsgs() {
		// try to extract the headerInfo and the client's status
		headerInfo, clientState := d.getHeaderAndClientState(ctx, msg)
		if headerInfo == nil {
			continue
		}

		// FrozenHeight is non-zero -> client is frozen -> this is a fork header
		// NOTE: A valid tx can ONLY have a single fork header msg, and this fork
		// header msg can ONLY be the LAST msg in this tx. If there is a fork
		// header before a canonical header in a tx, then the client will be
		// frozen upon the fork header, and the subsequent canonical header will
		// fail, eventually failing the entire tx. All state updates due to this
		// failed tx will be rolled back.
		isOnFork := !clientState.FrozenHeight.IsZero()
		d.k.HandleHeaderWithValidCommit(ctx, txHash, headerInfo, isOnFork)

		// unfreeze client (by setting FrozenHeight to zero again) if the client is frozen
		// due to a fork header
		if isOnFork {
			clientState.FrozenHeight = clienttypes.ZeroHeight()
			d.k.clientKeeper.SetClientState(ctx, headerInfo.ClientId, clientState)
		}
	}

	return next(ctx, tx, simulate, success)
}
