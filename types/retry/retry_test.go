package retry

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUnrecoverableError(t *testing.T) {
	err := Retry(1*time.Second, 1*time.Minute, func() error {
		return unrecoverableErrors[0]
	})
	require.Error(t, err)
}

func TestExpectedError(t *testing.T) {
	err := Retry(1*time.Second, 1*time.Minute, func() error {
		return expectedErrors[0]
	})
	require.NoError(t, err)
}
