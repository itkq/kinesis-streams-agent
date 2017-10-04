package sender

import (
	"testing"

	"github.com/itkq/kinesis-streams-agent/payload"
	"github.com/itkq/kinesis-streams-agent/sender/retry"
	"github.com/itkq/kinesis-streams-agent/state"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	sender := &Sender{
		client:        &SendOnlyOneRecordClient{},
		state:         &state.DummyState{},
		payloadCh:     make(chan *payload.Payload),
		backoff:       retry.NewExpBackOff(),
		RetryCountMax: 3,
	}

	records := []*payload.Record{
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
	}
	sender.SendWithRetry(records)
	m := sender.Export().(*SenderMetrics)
	assert.Equal(t, 0, m.RetryRecordsCount)
	assert.Equal(t, 0, len(m.RetryRecords))

	records = []*payload.Record{
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
	}
	sender.SendWithRetry(records)
	m = sender.Export().(*SenderMetrics)
	assert.Equal(t, 1, m.RetryRecordsCount)
	assert.Equal(t, 1, len(m.RetryRecords))
}
