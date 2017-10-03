package payloadbuffer

import (
	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/payload"
	"github.com/itkq/kinesis-agent-go/sender/kinesis"
)

const (
	DefaultRecordUnitSize    = kinesis.PutPayloadUnitSize
	DefaultRecordsPerPayload = kinesis.RecordCountMax
	DefaultPayloadSize       = kinesis.EntireRequestSizeMax
)

type PayloadBuffer struct {
	payload              *payload.Payload
	RecordUnitSize       int64
	RecordsPerPayloadMax int64
	PayloadSizeMax       int64
}

func NewPayloadBuffer() *PayloadBuffer {
	return &PayloadBuffer{
		payload:              payload.NewPayload(),
		RecordUnitSize:       DefaultRecordUnitSize,
		RecordsPerPayloadMax: DefaultRecordsPerPayload,
		PayloadSizeMax:       DefaultPayloadSize,
	}
}

func (b *PayloadBuffer) AddChunk(chunk *chunk.Chunk) *payload.Payload {
	lastRecord := b.payload.LastRecord()
	size := chunk.SendInfo.ReadRange.Len()

	// next payload (size over)
	if b.payload.Size+size > b.PayloadSizeMax {
		ret := b.Flush()

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.payload.AddRecord(r)

		return ret
	}

	// skip record aggregation
	if size > b.RecordUnitSize {
		var ret *payload.Payload = nil
		if b.payload.Count+1 > b.RecordsPerPayloadMax {
			ret = b.Flush()
		}

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.payload.AddRecord(r)

		return ret
	}

	// next record
	if lastRecord.Size+size > b.RecordUnitSize {
		var ret *payload.Payload = nil
		if b.payload.Count+1 > b.RecordsPerPayloadMax {
			ret = b.Flush()
		}

		r := payload.NewRecord()
		r.AddChunk(chunk)
		b.payload.AddRecord(r)

		return ret
	}

	// add chunk
	lastRecord.AddChunk(chunk)
	b.payload.Size += size

	return nil
}

func (b *PayloadBuffer) Flush() *payload.Payload {
	ret := *b.payload
	b.payload = payload.NewPayload()
	return &ret
}
