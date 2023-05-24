package types_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/stretchr/testify/require"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()

	ok := p1.Equal(p2)
	require.True(t, ok)

	p2.IbcPacketTimeoutSeconds = 100

	ok = p1.Equal(p2)
	require.False(t, ok)
}
