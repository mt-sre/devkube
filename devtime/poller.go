package devtime

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var DeadlineExceededErr = errors.New("set deadline exceeded")

type Poller struct {
	Clock           Clock
	MaxWaitDuration time.Duration
	PollInterval    time.Duration
}

func (p Poller) Wait(ctx context.Context, check func(context.Context) (bool, error)) error {
	clk := RealClockIfUnset(p.Clock)

	poll := p.PollInterval
	if poll == 0 {
		poll = 1 * time.Second
	}

	if p.MaxWaitDuration != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, p.MaxWaitDuration, DeadlineExceededErr)
		defer cancel()
	}

	ticker := clk.NewTicker(poll)
	tickerChan := ticker.C()
	defer ticker.Stop()

	slog.InfoContext(ctx, "waiting", "interval", poll)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tickerChan:
			done, err := check(ctx)
			switch {
			case err != nil:
				return err
			case done:
				return nil
			}
		}
	}
}
