package aggregator

import (
	"github.com/itkq/kinesis-streams-agent/payload"
)

func (a *Aggregator) Endpoint() string {
	return "/aggregator"
}

func (a *Aggregator) Export() interface{} {
	return &AggregatorMetrics{
		Payload: a.buffer.Payload,
	}
}

type AggregatorMetrics struct {
	Payload *payload.Payload `json:"payload"`
}
