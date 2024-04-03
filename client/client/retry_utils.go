package client

import (
	"cosmossdk.io/errors"
	"github.com/avast/retry-go/v4"
	"strings"
	"time"
)

// Variables used for retries
var (
	rtyAttNum = uint(5)
	rtyAtt    = retry.Attempts(rtyAttNum)
	rtyDel    = retry.Delay(time.Millisecond * 400)
	rtyErr    = retry.LastErrorOnly(true)
)

func errorContained(err error, errList []*errors.Error) bool {
	for _, e := range errList {
		if strings.Contains(err.Error(), e.Error()) {
			return true
		}
	}

	return false
}
