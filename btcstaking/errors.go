package btcstaking

import "errors"

var (
	ErrInvalidSlashingRate        = errors.New("invalid slashing rate")
	ErrDustOutputFound            = errors.New("transaction contains a dust output")
	ErrInsufficientSlashingAmount = errors.New("insufficient slashing amount")
	ErrInsufficientChangeAmount   = errors.New("insufficient change amount")
)
