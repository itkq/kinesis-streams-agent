package payloadbuffer

import (
	"fmt"
	"testing"

	"github.com/itkq/kinesis-agent-go/payload"
	"github.com/stretchr/testify/assert"

	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/state"
)

type testAddChunkCase struct {
	input           *chunk.Chunk
	expectedPayload *payload.Payload
	expectedSize    int64
	expectedCount   int64
	desc            string
}

var testAddChunkCases = []*testAddChunkCase{
	&testAddChunkCase{
		input: &chunk.Chunk{
			SendInfo: &state.SendInfo{
				Inode: 0,
				ReadRange: &state.FileReadRange{
					Begin: 0,
					End:   5,
				},
			},
			Body: []byte("hoge\n"),
		},
		expectedPayload: (*payload.Payload)(nil),
		expectedSize:    int64(5),
		expectedCount:   int64(1),
		desc:            "add chunk 1",
	},
	&testAddChunkCase{
		input: &chunk.Chunk{
			SendInfo: &state.SendInfo{
				Inode: 0,
				ReadRange: &state.FileReadRange{
					Begin: 5,
					End:   10,
				},
			},
			Body: []byte("fuga\n"),
		},
		expectedPayload: (*payload.Payload)(nil),
		expectedSize:    int64(10),
		expectedCount:   int64(1),
		desc:            "add chunk 2",
	},
	&testAddChunkCase{
		input: &chunk.Chunk{
			SendInfo: &state.SendInfo{
				Inode: 0,
				ReadRange: &state.FileReadRange{
					Begin: 10,
					End:   25,
				},
			},
			Body: []byte("hoge\nfuga\nhoge\n"),
		},
		expectedPayload: (*payload.Payload)(nil),
		expectedSize:    int64(25),
		expectedCount:   int64(2),
		desc:            "next record",
	},
	&testAddChunkCase{
		input: &chunk.Chunk{
			SendInfo: &state.SendInfo{
				Inode: 0,
				ReadRange: &state.FileReadRange{
					Begin: 25,
					End:   50,
				},
			},
			Body: []byte("fuga\nhoge\nfuga\nhoge\nfuga\n"),
		},
		expectedPayload: (*payload.Payload)(nil),
		expectedSize:    int64(50),
		expectedCount:   int64(3),
		desc:            "skip record aggregation",
	},
}

func TestAddChunk(t *testing.T) {
	buf := newTestPayloadBuffer()

	for _, c := range testAddChunkCases {
		p := buf.AddChunk(c.input)
		assert.Equal(t, c.expectedSize, buf.payload.Size, fmt.Sprintf("%s: size", c.desc))
		assert.Equal(t, c.expectedCount, buf.payload.Count, fmt.Sprintf("%s: count", c.desc))
		assert.Equal(t, c.expectedPayload, p, fmt.Sprintf("%s: payload", c.desc))
	}
}

func TestAddChunkWithFlush(t *testing.T) {
	expectedPayload := &payload.Payload{
		Size:  50,
		Count: 3,
		Records: []*payload.Record{
			&payload.Record{
				Size: 10,
				Chunks: []*chunk.Chunk{
					testAddChunkCases[0].input,
					testAddChunkCases[1].input,
				},
			},
			&payload.Record{
				Size: 15,
				Chunks: []*chunk.Chunk{
					testAddChunkCases[2].input,
				},
			},
			&payload.Record{
				Size: 25,
				Chunks: []*chunk.Chunk{
					testAddChunkCases[3].input,
				},
			},
		},
	}
	testAddChunkWithFlushCases := []*testAddChunkCase{
		&testAddChunkCase{
			input: &chunk.Chunk{
				SendInfo: &state.SendInfo{
					Inode: 10000,
					ReadRange: &state.FileReadRange{
						Begin: 50,
						End:   81,
					},
				},
				Body: []byte("dummy\n"),
			},
			expectedPayload: expectedPayload,
			expectedSize:    int64(30),
			expectedCount:   int64(1),
			desc:            "total size over",
		},
		&testAddChunkCase{
			input: &chunk.Chunk{
				SendInfo: &state.SendInfo{
					Inode: 10000,
					ReadRange: &state.FileReadRange{
						Begin: 50,
						End:   71,
					},
				},
				Body: []byte("dummy\n"),
			},
			expectedPayload: expectedPayload,
			expectedSize:    int64(21),
			expectedCount:   int64(1),
			desc:            "recordUnitSize over",
		},
		&testAddChunkCase{
			input: &chunk.Chunk{
				SendInfo: &state.SendInfo{
					Inode: 10000,
					ReadRange: &state.FileReadRange{
						Begin: 50,
						End:   60,
					},
				},
				Body: []byte("dummy\n"),
			},
			expectedPayload: expectedPayload,
			expectedSize:    int64(10),
			expectedCount:   int64(1),
			desc:            "record size over",
		},
	}

	for _, c := range testAddChunkWithFlushCases {
		buf := newTestPayloadBuffer()

		for _, cc := range testAddChunkCases {
			buf.AddChunk(cc.input)
		}

		p := buf.AddChunk(c.input)

		for i, _ := range p.Records {
			for j, _ := range p.Records[i].Chunks {
				assert.Equal(
					t,
					c.expectedPayload.Records[i].Chunks[j].Body,
					p.Records[i].Chunks[j].Body,
				)
				assert.Equal(
					t,
					*c.expectedPayload.Records[i].Chunks[j].SendInfo,
					*p.Records[i].Chunks[j].SendInfo,
				)
			}
		}
	}
}

func newTestPayloadBuffer() *PayloadBuffer {
	buf := NewPayloadBuffer()
	buf.RecordUnitSize = 20
	buf.RecordsPerPayloadMax = 3
	buf.PayloadSizeMax = 80

	return buf
}
