package devtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mt-sre/devkube/devtime"

	"github.com/stretchr/testify/require"
)

func TestPollerNothingSet(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time)
	close(tc)
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 1*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock}
	called := 0
	ctx := context.Background()
	check := func(c context.Context) (bool, error) {
		require.Equal(t, ctx, c)
		called++
		return true, nil
	}
	err := poller.Wait(ctx, check)
	require.NoError(t, err)
	require.Equal(t, 1, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}

func TestPollerIntervalSet(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time)
	close(tc)
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 7*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock, PollInterval: 7 * time.Second}
	called := 0
	ctx := context.Background()
	check := func(c context.Context) (bool, error) {
		require.Equal(t, ctx, c)
		called++
		return true, nil
	}
	err := poller.Wait(ctx, check)
	require.NoError(t, err)
	require.Equal(t, 1, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}

func TestPollerCheckErrors(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time)
	close(tc)
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 1*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock}
	called := 0
	ctx := context.Background()
	checkErr := errors.New("ogno")
	check := func(c context.Context) (bool, error) {
		require.Equal(t, ctx, c)
		called++
		return true, checkErr
	}
	err := poller.Wait(ctx, check)
	require.ErrorIs(t, err, checkErr)
	require.Equal(t, 1, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}

func TestPollerMultipoll(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time)
	close(tc)
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 1*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock}
	called := 0
	ctx := context.Background()
	check := func(c context.Context) (bool, error) {
		require.Equal(t, ctx, c)
		called++
		if called > 2 {
			return true, nil
		}
		return false, nil
	}
	err := poller.Wait(ctx, check)
	require.NoError(t, err)
	require.Equal(t, 3, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}

func TestPollerCanceledContext(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time, 1)
	tc <- time.Time{}
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 1*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock}

	ctx, cancel := context.WithCancel(context.Background())
	called := 0
	check := func(c context.Context) (bool, error) {
		require.Equal(t, ctx, c)
		cancel()
		called++
		return false, nil
	}
	err := poller.Wait(ctx, check)
	require.ErrorIs(t, err, ctx.Err())
	require.Equal(t, 1, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}

func TestPollerMaxWait(t *testing.T) {
	t.Parallel()

	clock := &devtime.MockClock{}
	ticker := &devtime.MockTicker{}
	tc := make(chan time.Time, 1)
	tc <- time.Time{}
	ticker.On("C").Once().Return(tc)
	ticker.On("Stop").Once()
	clock.On("NewTicker", 1*time.Second).Once().Return(ticker)
	poller := devtime.Poller{Clock: clock, MaxWaitDuration: 6 * time.Second}

	ctx := context.Background()
	called := 0
	check := func(c context.Context) (bool, error) {
		require.NotEqual(t, ctx, c)
		called++
		return true, nil
	}
	err := poller.Wait(ctx, check)
	require.ErrorIs(t, err, ctx.Err())
	require.Equal(t, 1, called)

	clock.AssertExpectations(t)
	ticker.AssertExpectations(t)
}
