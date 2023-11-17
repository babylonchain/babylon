package types

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func CheckForDuplicatesAndEmptyStrings(input []string) error {
	encountered := map[string]bool{}
	for _, str := range input {
		if len(str) == 0 {
			return fmt.Errorf("empty string is not allowed")
		}

		if encountered[str] {
			return fmt.Errorf("duplicate entry found: %s", str)
		}

		encountered[str] = true
	}

	return nil
}

// IsValidSlashingRate checks if the given slashing rate is b/w the valid range i.e., (0,1)
func IsValidSlashingRate(slashingRate sdk.Dec) bool {
	// TODO: add check to confirm precision is max 2 decimal places

	return slashingRate.GT(sdk.ZeroDec()) && slashingRate.LT(sdk.OneDec())
}
