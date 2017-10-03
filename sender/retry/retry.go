package retry

import (
	"time"
)

type BackOff interface {
	NextBackOff() time.Duration
	Reset()
	GetRetryCount() int
}

func Retry(n int, b BackOff, fn func() error) error {
	var err error

	b.Reset()
	for i := n; i > 0; i-- {
		err = fn()
		if err == nil {
			break
		}

		time.Sleep(b.NextBackOff())
	}

	return err
}
