package retry

import (
	"math/rand"
	"time"
)

const (
	DefaultInitialInterval     = 100 * time.Millisecond
	DefaultMultiplier          = 2
	DefaultRandomizationFactor = 0.1
)

// exponential backoff
type ExpBackOff struct {
	InitialInterval     time.Duration
	Multiplier          float64
	RandomizationFactor float64

	currentInterval time.Duration
	random          *rand.Rand
	retryCount      int
}

func NewExpBackOff() *ExpBackOff {
	return &ExpBackOff{
		InitialInterval:     DefaultInitialInterval,
		Multiplier:          DefaultMultiplier,
		RandomizationFactor: DefaultRandomizationFactor,
		currentInterval:     DefaultInitialInterval,
		random:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *ExpBackOff) NextBackOff() time.Duration {
	defer b.incrementCurrentInterval()
	if b.random == nil {
		b.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return getRandomValueFromInterval(b.RandomizationFactor, b.random.Float64(), b.currentInterval)
}

func (b *ExpBackOff) Reset() {
	b.currentInterval = b.InitialInterval
	b.retryCount = 0
}

func (b *ExpBackOff) GetRetryCount() int {
	return b.retryCount
}

// https://github.com/cenkalti/backoff/blob/master/exponential.go
// Increments the current interval by multiplying it with the multiplier.
func (b *ExpBackOff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	b.currentInterval = time.Duration(float64(b.currentInterval) * b.Multiplier)
	b.retryCount++
}

// https://github.com/cenkalti/backoff/blob/master/exponential.go
// Returns a random value from the following interval:
// 	[randomizationFactor * currentInterval, randomizationFactor * currentInterval].
func getRandomValueFromInterval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	var delta = randomizationFactor * float64(currentInterval)
	var minInterval = float64(currentInterval) - delta
	var maxInterval = float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}
