package aggregator

import (
	"testing"
	"time"

	"github.com/itkq/kinesis-streams-agent/state"
	"github.com/stretchr/testify/assert"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/payload"

	"github.com/itkq/kinesis-streams-agent/aggregator/payload_buffer"
)

func TestAggregatorRun(t *testing.T) {
	buffer := payloadbuffer.NewPayloadBuffer()
	buffer.RecordUnitSize = 10
	buffer.RecordsPerPayloadMax = 3
	buffer.PayloadSizeMax = 50
	aggr := &Aggregator{
		buffer:    buffer,
		ChunkCh:   make(chan *chunk.Chunk),
		PayloadCh: make(chan *payload.Payload),
	}
	aggr.FlushInterval = 200 * time.Millisecond

	go aggr.Run()

	aggr.ChunkCh <- &chunk.Chunk{
		SendInfo: &state.SendInfo{
			ReadRange: &state.FileReadRange{
				Begin: 0,
				End:   5,
			},
		},
		Body: []byte("dummy\n"),
	}
	aggr.ChunkCh <- &chunk.Chunk{
		SendInfo: &state.SendInfo{
			ReadRange: &state.FileReadRange{
				Begin: 5,
				End:   55,
			},
		},
		Body: []byte("dummy\n"),
	}

	// aggregate and send
	p := <-aggr.PayloadCh
	assert.Equal(t, int64(5), p.Size)

	// interval flush send
	time.Sleep(400 * time.Millisecond)
	p = <-aggr.PayloadCh
	assert.Equal(t, int64(50), p.Size)
}
