package devclock_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/devkube/devclock"
)

func TestReadClockIfUnset(t *testing.T) {
	t.Parallel()

	mockClock := &devclock.MockClock{}

	res := devclock.RealClockIfUnset(mockClock)
	require.Equal(t, mockClock, res)

	res = devclock.RealClockIfUnset(nil)
	require.IsType(t, devclock.RealClock{}, res)
}
