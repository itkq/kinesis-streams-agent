package retry

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpBackoff(t *testing.T) {
	b := NewExpBackOff()

	for i := 1; i < 3; i++ {
		currentInterval := b.currentInterval
		interval := b.NextBackOff()
		assert.Equal(t, i, b.GetRetryCount())
		assert.True(
			t,
			inFactorRange(interval, currentInterval, b.RandomizationFactor),
		)
		assert.Equal(
			t,
			b.InitialInterval*time.Duration(math.Pow(b.Multiplier, float64(i))),
			b.currentInterval,
		)
	}

	b.Reset()
	assert.Equal(t, b.InitialInterval, b.currentInterval)
}

func inFactorRange(i, compare time.Duration, factor float64) bool {
	f := float64(compare)
	return time.Duration(f-f*factor) <= i && i <= time.Duration(f+f*factor)
}
