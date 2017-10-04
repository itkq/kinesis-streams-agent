package chunk

import (
	"github.com/itkq/kinesis-streams-agent/state"
)

type Chunk struct {
	SendInfo *state.SendInfo
	Body     []byte
}
