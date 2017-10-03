package sender

import (
	"encoding/json"

	"github.com/itkq/kinesis-agent-go/api"
	"github.com/itkq/kinesis-agent-go/payload"
)

func (s *Sender) Endpoint() string {
	return "/sender"
}

func (s *Sender) Export() api.Metrics {
	return &SenderMetrics{
		RetryRecords:      s.retryRecords,
		RetryRecordsCount: len(s.retryRecords),
	}
}

type SenderMetrics struct {
	RetryRecords      []*payload.Record
	RetryRecordsCount int
}

func (s *SenderMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}
