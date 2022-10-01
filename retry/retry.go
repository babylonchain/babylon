package retry

import (
	"errors"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"math/rand"
	"time"
)

// unrecoverableErrors is a list of errors which are unsafe and should not be retried.
var unrecoverableErrors = []error{
	btclctypes.ErrHeaderParentDoesNotExist,
	btcctypes.ErrProvidedHeaderDoesNotHaveAncestor,
	btcctypes.ErrUnknownHeader,
	btcctypes.ErrNoCheckpointsForPreviousEpoch,
	btcctypes.ErrInvalidCheckpointProof,
	// TODO Add more errors here
}

// expectedErrors is a list of errors which can safely be ignored and should not be retried.
var expectedErrors = []error{
	btclctypes.ErrDuplicateHeader,
	btcctypes.ErrDuplicatedSubmission,
	btcctypes.ErrUnknownHeader,
	// TODO Add more errors here
}

func isUnrecoverableErr(err error) bool {
	for _, e := range unrecoverableErrors {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}

func isExpectedErr(err error) bool {
	for _, e := range expectedErrors {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}

func Retry(sleep time.Duration, maxSleepTime time.Duration, retryableFunc func() error) error {
	if err := retryableFunc(); err != nil {
		if isUnrecoverableErr(err) {
			log.Warnf("Skip retry, error unrecoverable %v", err)
			return err
		}

		if isExpectedErr(err) {
			log.Warnf("Skip retry, error expected %v", err)
			return nil
		}

		// Add some randomness to prevent thrashing
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		sleep = sleep + jitter/2

		if sleep > maxSleepTime {
			log.Info("retry timed out")
			return err
		}

		log.Warnf("sleeping for %v sec: %v", sleep, err)
		time.Sleep(sleep)

		return Retry(2*sleep, maxSleepTime, retryableFunc)
	}
	return nil
}
