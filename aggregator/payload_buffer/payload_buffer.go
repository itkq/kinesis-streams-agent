package payloadbuffer

import (
	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/payload"
	"github.com/itkq/kinesis-streams-agent/sender/kinesis"
)

const (
	DefaultRecordUnitSize    = kinesis.PutPayloadUnitSize
	DefaultRecordsPerPayload = kinesis.RecordCountMax
	DefaultPayloadSize       = kinesis.EntireRequestSizeMax
)

type PayloadBuffer struct {
	Payload              *payload.Payload
	RecordUnitSize       int64
	RecordsPerPayloadMax int64
	PayloadSizeMax       int64
}

func NewPayloadBuffer() *PayloadBuffer {
	return &PayloadBuffer{
		Payload:              payload.NewPayload(),
		RecordUnitSize:       DefaultRecordUnitSize,
		RecordsPerPayloadMax: DefaultRecordsPerPayload,
		PayloadSizeMax:       DefaultPayloadSize,
	}
}

func (b *PayloadBuffer) AddChunk(chunk *chunk.Chunk) *payload.Payload {
	lastRecord := b.Payload.LastRecord()
	size := chunk.SendInfo.ReadRange.Len()

	// next payload (size over)
	if b.Payload.Size+size > b.PayloadSizeMax {
		ret := b.Flush()

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.Payload.AddRecord(r)

		return ret
	}

	// skip record aggregation
	if size > b.RecordUnitSize {
		var ret *payload.Payload = nil
		if b.Payload.Count+1 > b.RecordsPerPayloadMax {
			ret = b.Flush()
		}

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.Payload.AddRecord(r)

		return ret
	}

	// next record
	if lastRecord.Size+size > b.RecordUnitSize {
		var ret *payload.Payload = nil
		if b.Payload.Count+1 > b.RecordsPerPayloadMax {
			ret = b.Flush()
		}

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.Payload.AddRecord(r)

		return ret
	}

	// add chunk
	lastRecord.AddChunk(chunk)
	b.Payload.Size += size

	return nil
}

func (b *PayloadBuffer) Flush() *payload.Payload {
	ret := *b.Payload
	b.Payload = payload.NewPayload()
	return &ret
}
