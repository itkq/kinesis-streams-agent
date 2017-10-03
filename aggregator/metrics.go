package aggregator

import (
	"encoding/json"

	"github.com/itkq/kinesis-agent-go/api"
	"github.com/itkq/kinesis-agent-go/payload"
)

func (a *Aggregator) Endpoint() string {
	return "/aggregator"
}

func (a *Aggregator) Export() api.Metrics {
	return &AggregatorMetrics{
		Payload: a.buffer.Payload,
	}
}

type AggregatorMetrics struct {
	Payload *payload.Payload `json:"payload"`
}

func (m *AggregatorMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
