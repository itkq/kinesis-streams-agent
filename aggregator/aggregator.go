package aggregator

import (
	"log"
	"time"

	"github.com/itkq/kinesis-agent-go/aggregator/payload_buffer"
	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/config"
	"github.com/itkq/kinesis-agent-go/payload"
)

type Aggregator struct {
	// input channel
	ChunkCh chan *chunk.Chunk
	// output channel
	PayloadCh chan *payload.Payload

	buffer      *payloadbuffer.PayloadBuffer
	flushTicker *time.Ticker
}

func NewAggregator(c *config.AggregatorConfig) *Aggregator {
	return &Aggregator{
		buffer:      payloadbuffer.NewPayloadBuffer(),
		ChunkCh:     make(chan *chunk.Chunk),
		PayloadCh:   make(chan *payload.Payload),
		flushTicker: time.NewTicker(c.FlushInverval),
	}
}

func (a *Aggregator) Run() {
	for {
		select {
		case chunk := <-a.ChunkCh:
			p := a.Aggregate(chunk)
			a.Output(p)

		case <-a.flushTicker.C:
			log.Println("aggregator> interval flush")
			a.Flush()
		}
	}
}

func (a *Aggregator) Aggregate(chunk *chunk.Chunk) *payload.Payload {
	return a.buffer.AddChunk(chunk)
}

func (a *Aggregator) Output(p *payload.Payload) {
	if p != nil && p.Size > 0 {
		a.PayloadCh <- p
	}
}

func (a *Aggregator) Flush() {
	p := a.buffer.Flush()
	a.Output(p)
}
