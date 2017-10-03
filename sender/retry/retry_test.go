package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DummyBackOff struct {
	cnt int
}

func (b *DummyBackOff) NextBackOff() time.Duration {
	b.cnt++
	return 1 * time.Nanosecond
}

func (b *DummyBackOff) Reset() {
	b.cnt = 0
}

func (b *DummyBackOff) GetRetryCount() int {
	return b.cnt
}

func TestRetrySuccess(t *testing.T) {
	cnt := 0

	err := Retry(3, &DummyBackOff{}, func() error {
		if cnt == 0 {
			cnt++
			return fmt.Errorf("retry")
		}

		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, cnt)
}

func TestRetryFailed(t *testing.T) {
	cnt := 0
	err := Retry(4, &DummyBackOff{}, func() error {
		cnt++
		return fmt.Errorf("retry")
	})

	assert.Error(t, err)
	assert.Equal(t, 4, cnt)
}
