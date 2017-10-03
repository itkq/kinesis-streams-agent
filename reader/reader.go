package reader

import (
	"time"

	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/state"
)

type Reader interface {
	Run(readerState *state.ReaderState, clockCh <-chan time.Time, chunkCh chan<- *chunk.Chunk)
	Opened() bool
}
