package payload

import "github.com/itkq/kinesis-streams-agent/chunk"

type Record struct {
	Size         int64
	Chunks       []*chunk.Chunk
	ErrorCode    *string
	ErrorMessage *string
}

func NewRecord() *Record {
	return &Record{
		Chunks: make([]*chunk.Chunk, 0),
	}
}

func (r *Record) AddChunk(chunk *chunk.Chunk) {
	r.Chunks = append(r.Chunks, chunk)
	r.Size += chunk.SendInfo.ReadRange.Len()
}

func (r *Record) Success() {
	for _, c := range r.Chunks {
		c.SendInfo.Succeeded = true
	}
}

func (r *Record) ToByte() []byte {
	var b []byte
	for _, chunk := range r.Chunks {
		b = append(b, chunk.Body...)
	}

	return b
}
