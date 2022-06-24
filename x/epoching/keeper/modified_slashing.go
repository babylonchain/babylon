package keeper

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// HandleValidatorSignature handles a validator signature (for slashing equivocating validators)
// called once per validator per block.
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/slashing/keeper/infractions.go#L11-L126)
func HandleValidatorSignature(ctx sdk.Context, slk *slashingkeeper.Keeper, stk *stakingkeeper.Keeper, addr cryptotypes.Address, power int64, signed bool) {
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
			// Downtime confirmed: slash the validator
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
			// TODO: commented the line below as it's a private function. Find a way to work around this.
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
