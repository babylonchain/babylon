package types

import (
	"fmt"
	"reflect"
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
			return fmt.Errorf("string cannot be empty")
		}

		if encountered[str] {
			return fmt.Errorf("duplicate string: %s", str)
		}

		encountered[str] = true
	}

	return nil
}
