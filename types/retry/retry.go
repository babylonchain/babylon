package retry

import (
	"errors"
	"math/rand"
	"time"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

// unrecoverableErrors is a list of errors which are unsafe and should not be retried.
var unrecoverableErrors = []error{
	btclctypes.ErrHeaderParentDoesNotExist,
	btcctypes.ErrProvidedHeaderDoesNotHaveAncestor,
	btcctypes.ErrInvalidHeader,
	btcctypes.ErrNoCheckpointsForPreviousEpoch,
	btcctypes.ErrInvalidCheckpointProof,
	checkpointingtypes.ErrBlsPrivKeyDoesNotExist,
	// TODO Add more errors here
}

// expectedErrors is a list of errors which can safely be ignored and should not be retried.
var expectedErrors = []error{
	btclctypes.ErrDuplicateHeader,
	btcctypes.ErrDuplicatedSubmission,
	btcctypes.ErrInvalidHeader,
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

func Do(sleep time.Duration, maxSleepTime time.Duration, retryableFunc func() error) error {
	if err := retryableFunc(); err != nil {
		if isUnrecoverableErr(err) {
			logger.Error("Skip retry, error unrecoverable", "err", err)
			return err
		}

		if isExpectedErr(err) {
			logger.Error("Skip retry, error expected", "err", err)
			return nil
		}

		// Add some randomness to prevent thrashing
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		sleep = sleep + jitter/2

		if sleep > maxSleepTime {
			logger.Info("retry timed out")
			return err
		}

		logger.Info("starting exponential backoff", "sleep", sleep, "err", err)
		time.Sleep(sleep)

		return Do(2*sleep, maxSleepTime, retryableFunc)
	}
	return nil
}
