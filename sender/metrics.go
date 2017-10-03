package sender

import (
	"github.com/itkq/kinesis-agent-go/payload"
)

func (s *Sender) Endpoint() string {
	return "/sender"
}

func (s *Sender) Export() interface{} {
	return &SenderMetrics{
		RetryRecords:      s.retryRecords,
		RetryRecordsCount: len(s.retryRecords),
	}
}

type SenderMetrics struct {
	RetryRecords      []*payload.Record
	RetryRecordsCount int
}
