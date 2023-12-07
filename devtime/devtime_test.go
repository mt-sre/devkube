package devtime_test

import (
	"testing"

	"github.com/mt-sre/devkube/devtime"

	"github.com/stretchr/testify/require"
)

func TestReadClockIfUnset(t *testing.T) {
	t.Parallel()

	mockClock := &devtime.MockClock{}

	res := devtime.RealClockIfUnset(mockClock)
	require.Equal(t, mockClock, res)

	res = devtime.RealClockIfUnset(nil)
	require.IsType(t, devtime.RealClock{}, res)
}
