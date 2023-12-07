//nolint:revive
package devtime

import (
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/utils/clock"
)

type MockClock struct{ mock.Mock }

func (c *MockClock) NewTicker(d time.Duration) clock.Ticker {
	return c.Called(d).Get(0).(clock.Ticker)
}

type MockTicker struct{ mock.Mock }

func (t *MockTicker) C() <-chan time.Time { return t.Called().Get(0).(chan time.Time) }
func (t *MockTicker) Stop()               { t.Called() }
