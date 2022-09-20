package types

import (
	"math/rand"
	"reflect"
	"time"
)

func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func Retry(sleep time.Duration, maxSleepTime time.Duration, f func() error) error {
	if err := f(); err != nil {
		// Add some randomness to prevent thrashing
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		sleep = sleep + jitter/2

		if sleep > maxSleepTime {
			return err
		}

		time.Sleep(sleep)

		return Retry(2*sleep, maxSleepTime, f)
	}
	return nil
}
