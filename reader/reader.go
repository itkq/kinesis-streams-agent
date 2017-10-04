package reader

import (
	"time"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/state"
)

type Reader interface {
	Run(readerState *state.ReaderState, clockCh <-chan time.Time, chunkCh chan<- *chunk.Chunk)
	Opened() bool
}
