package aggregator

import (
	"testing"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/state"
	"github.com/stretchr/testify/assert"
)

func TestExport(t *testing.T) {
	aggr := NewAggregator()
	aggr.Aggregate(&chunk.Chunk{
		Body: []byte("hoge"),
		SendInfo: &state.SendInfo{
			ReadRange: &state.FileReadRange{
				Begin: 0,
				End:   4,
			},
		},
	})

	assert.Equal(t, "/aggregator", aggr.Endpoint())
	p := aggr.Export().(*AggregatorMetrics).Payload
	assert.Equal(t, int64(1), p.Count)
	assert.Equal(t, int64(4), p.Size)
}
