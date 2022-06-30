package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeKey             sdk.StoreKey
		memKey               sdk.StoreKey
		paramstore           paramtypes.Subspace
		btcLightClientKeeper types.BTCLightClientKeeper
		checkpointingKeeper  types.CheckpointingKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	bk types.BTCLightClientKeeper,
	ck types.CheckpointingKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		paramstore:           ps,
		btcLightClientKeeper: bk,
		checkpointingKeeper:  ck,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetBlockHeight(b chainhash.Hash) (uint64, error) {
	return k.btcLightClientKeeper.BlockHeight(b)
}

func (k Keeper) IsAncestor(parentHash chainhash.Hash, childHash chainhash.Hash) (bool, error) {
	return k.btcLightClientKeeper.IsAncestor(parentHash, childHash)
}

func (k Keeper) GetCheckpointEpoch(c []byte) (uint64, error) {
	return k.checkpointingKeeper.CheckpointEpoch(c)
}

// TODO for now we jsut store raw checkpoint with epoch as key. Ultimatly
// we should store checkpoint with some timestamp info, or even do not store
// checkpoint itelf but its status
func (k Keeper) StoreCheckpoint(ctx sdk.Context, e uint64, c []byte) {
	store := ctx.KVStore(k.storeKey)
	key := sdk.Uint64ToBigEndian(e)
	store.Set(key, c)
}

// TODO just return checkpoint if exists
func (k Keeper) GetCheckpoint(ctx sdk.Context, e uint64) []byte {
	store := ctx.KVStore(k.storeKey)
	key := sdk.Uint64ToBigEndian(e)
	return store.Get(key)
}

func (k Keeper) SubmissionExists(ctx sdk.Context, sk types.SubmissionKey) bool {
	store := ctx.KVStore(k.storeKey)
	kBytes := k.cdc.MustMarshal(&sk)
	return store.Has(kBytes)
}

// Return epoch data for given epoch, if there is not epoch data yet returns nil
func (k Keeper) GetEpochData(ctx sdk.Context, e uint64) *types.EpochData {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetEpochIndexKey(e))
	if len(bytes) == 0 {
		return nil
	} else {
		ed := &types.EpochData{}
		k.cdc.MustUnmarshal(bytes, ed)
		return ed
	}
}

func (k Keeper) SaveEpochData(ctx sdk.Context, e uint64, ed *types.EpochData) {
	store := ctx.KVStore(k.storeKey)
	ek := types.GetEpochIndexKey(e)
	eb := k.cdc.MustMarshal(ed)
	store.Set(ek, eb)
}

func (k Keeper) SaveSubmission(ctx sdk.Context, sk types.SubmissionKey, sd types.SubmissionData) {
	store := ctx.KVStore(k.storeKey)
	kBytes := k.cdc.MustMarshal(&sk)
	sBytes := k.cdc.MustMarshal(&sd)
	store.Set(kBytes, sBytes)
}
