package keeper

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// HandleValidatorSignature handles a validator signature (for slashing equivocating validators)
// called once per validator per block.
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/slashing/keeper/infractions.go#L11-L126)
func HandleValidatorSignature(ctx sdk.Context, slk slashingkeeper.Keeper, stk stakingkeeper.Keeper, addr cryptotypes.Address, power int64, signed bool) {
	logger := slk.Logger(ctx)
	height := ctx.BlockHeight()

	// fetch the validator public key
	consAddr := sdk.ConsAddress(addr)
	if _, err := slk.GetPubkey(ctx, addr); err != nil {
		panic(fmt.Sprintf("Validator consensus-address %s not found", consAddr))
	}

	// fetch signing info
	signInfo, found := slk.GetValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}

	// this is a relative index, so it counts blocks the validator *should* have signed
	// will use the 0-value default signing info if not present, except for start height
	index := signInfo.IndexOffset % slk.SignedBlocksWindow(ctx)
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := slk.GetValidatorMissedBlockBitArray(ctx, consAddr, index)
	missed := !signed
	switch {
	case !previous && missed:
		// Array value has changed from not missed to missed, increment counter
		slk.SetValidatorMissedBlockBitArray(ctx, consAddr, index, true)
		signInfo.MissedBlocksCounter++
	case previous && !missed:
		// Array value has changed from missed to not missed, decrement counter
		slk.SetValidatorMissedBlockBitArray(ctx, consAddr, index, false)
		signInfo.MissedBlocksCounter--
	default:
		// Array value at this index has not changed, no need to update counter
	}

	minSignedPerWindow := slk.MinSignedPerWindow(ctx)

	if missed {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				slashingtypes.EventTypeLiveness,
				sdk.NewAttribute(slashingtypes.AttributeKeyAddress, consAddr.String()),
				sdk.NewAttribute(slashingtypes.AttributeKeyMissedBlocks, fmt.Sprintf("%d", signInfo.MissedBlocksCounter)),
				sdk.NewAttribute(slashingtypes.AttributeKeyHeight, fmt.Sprintf("%d", height)),
			),
		)

		logger.Debug(
			"absent validator",
			"height", height,
			"validator", consAddr.String(),
			"missed", signInfo.MissedBlocksCounter,
			"threshold", minSignedPerWindow,
		)
	}

	minHeight := signInfo.StartHeight + slk.SignedBlocksWindow(ctx)
	maxMissed := slk.SignedBlocksWindow(ctx) - minSignedPerWindow

	// if we are past the minimum height and the validator has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		validator := stk.ValidatorByConsAddr(ctx, consAddr)
		if validator != nil && !validator.IsJailed() {
			// Downtime confirmed: slash and jail the validator
			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - sdk.ValidatorUpdateDelay - 1

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					slashingtypes.EventTypeSlash,
					sdk.NewAttribute(slashingtypes.AttributeKeyAddress, consAddr.String()),
					sdk.NewAttribute(slashingtypes.AttributeKeyPower, fmt.Sprintf("%d", power)),
					sdk.NewAttribute(slashingtypes.AttributeKeyReason, slashingtypes.AttributeValueMissingSignature),
					// sdk.NewAttribute(slashingtypes.AttributeKeyJailed, consAddr.String()),
				),
			)
			stk.Slash(ctx, consAddr, distributionHeight, power, slk.SlashFractionDowntime(ctx))
			// stk.Jail(ctx, consAddr)

			// signInfo.JailedUntil = ctx.BlockHeader().Time.Add(slk.DowntimeJailDuration(ctx))

			// We need to reset the counter & array so that the validator won't be immediately slashed for downtime upon rebonding.
			signInfo.MissedBlocksCounter = 0
			signInfo.IndexOffset = 0
			// slk.clearValidatorMissedBlockBitArray(ctx, consAddr)

			logger.Info(
				// "slashing and jailing validator due to liveness fault",
				"slashing validator due to liveness fault",
				"height", height,
				"validator", consAddr.String(),
				"min_height", minHeight,
				"threshold", minSignedPerWindow,
				"slashed", slk.SlashFractionDowntime(ctx).String(),
				// "jailed_until", signInfo.JailedUntil,
			)
		} else {
			// validator was (a) not found or (b) already jailed so we do not slash
			logger.Info(
				"validator would have been slashed for downtime, but was either not found in store or already jailed",
				"validator", consAddr.String(),
			)
		}
	}

	// Set the updated signing info
	slk.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

// HandleEquivocationEvidence implements an equivocation evidence handler. Assuming the
// evidence is valid, the validator committing the misbehavior will be slashed,
// jailed and tombstoned. Once tombstoned, the validator will not be able to
// recover. Note, the evidence contains the block time and height at the time of
// the equivocation.
//
// The evidence is considered invalid if:
// - the evidence is too old
// - the validator is unbonded or does not exist
// - the signing info does not exist (will panic)
// - is already tombstoned
//
// TODO: Some of the invalid constraints listed above may need to be reconsidered
// in the case of a lunatic attack.
func HandleEquivocationEvidence(ctx sdk.Context, ek *evidencekeeper.Keeper, slk *slashingkeeper.Keeper, stk *stakingkeeper.Keeper, evidence *evidencetypes.Equivocation) {
	logger := ek.Logger(ctx)
	consAddr := evidence.GetConsensusAddress()

	if _, err := slk.GetPubkey(ctx, consAddr.Bytes()); err != nil {
		// Ignore evidence that cannot be handled.
		//
		// NOTE: We used to panic with:
		// `panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))`,
		// but this couples the expectations of the app to both Tendermint and
		// the simulator.  Both are expected to provide the full range of
		// allowable but none of the disallowed evidence types.  Instead of
		// getting this coordination right, it is easier to relax the
		// constraints and ignore evidence that cannot be handled.
		return
	}

	// calculate the age of the evidence
	infractionHeight := evidence.GetHeight()
	infractionTime := evidence.GetTime()
	ageDuration := ctx.BlockHeader().Time.Sub(infractionTime)
	ageBlocks := ctx.BlockHeader().Height - infractionHeight

	// Reject evidence if the double-sign is too old. Evidence is considered stale
	// if the difference in time and number of blocks is greater than the allowed
	// parameters defined.
	cp := ctx.ConsensusParams()
	if cp != nil && cp.Evidence != nil {
		if ageDuration > cp.Evidence.MaxAgeDuration && ageBlocks > cp.Evidence.MaxAgeNumBlocks {
			logger.Info(
				"ignored equivocation; evidence too old",
				"validator", consAddr,
				"infraction_height", infractionHeight,
				"max_age_num_blocks", cp.Evidence.MaxAgeNumBlocks,
				"infraction_time", infractionTime,
				"max_age_duration", cp.Evidence.MaxAgeDuration,
			)
			return
		}
	}

	validator := stk.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.IsUnbonded() {
		// Defensive: Simulation doesn't take unbonding periods into account, and
		// Tendermint might break this assumption at some point.
		return
	}

	if ok := slk.HasValidatorSigningInfo(ctx, consAddr); !ok {
		panic(fmt.Sprintf("expected signing info for validator %s but not found", consAddr))
	}

	// ignore if the validator is already tombstoned
	if slk.IsTombstoned(ctx, consAddr) {
		logger.Info(
			"ignored equivocation; validator already tombstoned",
			"validator", consAddr,
			"infraction_height", infractionHeight,
			"infraction_time", infractionTime,
		)
		return
	}

	logger.Info(
		"confirmed equivocation",
		"validator", consAddr,
		"infraction_height", infractionHeight,
		"infraction_time", infractionTime,
	)

	// We need to retrieve the stake distribution which signed the block, so we
	// subtract ValidatorUpdateDelay from the evidence height.
	// Note, that this *can* result in a negative "distributionHeight", up to
	// -ValidatorUpdateDelay, i.e. at the end of the
	// pre-genesis block (none) = at the beginning of the genesis block.
	// That's fine since this is just used to filter unbonding delegations & redelegations.
	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	// Slash validator. The `power` is the int64 power of the validator as provided
	// to/by Tendermint. This value is validator.Tokens as sent to Tendermint via
	// ABCI, and now received as evidence. The fraction is passed in to separately
	// to slash unbonding and rebonding delegations.
	slk.Slash(
		ctx,
		consAddr,
		slk.SlashFractionDoubleSign(ctx),
		evidence.GetValidatorPower(), distributionHeight,
	)

	// // Jail the validator if not already jailed. This will begin unbonding the
	// // validator if not already unbonding (tombstoned).
	// if !validator.IsJailed() {
	// 	slk.Jail(ctx, consAddr)
	// }

	// slk.JailUntil(ctx, consAddr, evidencetypes.DoubleSignJailEndTime)
	slk.Tombstone(ctx, consAddr)
	ek.SetEvidence(ctx, evidence)
}
