package types_test

import (
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func TestValidatorSet_FindValidatorWithIndex(t *testing.T) {
	valSet1 := datagen.GenRandomValSet(10)
	valSet2 := datagen.GenRandomValSet(1)
	for i := 0; i < len(valSet1); i++ {
		val, index, err := valSet1.FindValidatorWithIndex(valSet1[i].Addr)
		require.NoError(t, err)
		require.Equal(t, val, &valSet1[i])
		require.Equal(t, index, i)
	}
	val, index, err := valSet1.FindValidatorWithIndex(valSet2[0].Addr)
	require.Error(t, err)
	require.Nil(t, val)
	require.Equal(t, 0, index)
}
