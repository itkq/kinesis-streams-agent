package chunk

import (
	"github.com/itkq/kinesis-agent-go/state"
)

type Chunk struct {
	SendInfo *state.SendInfo
	Body     []byte
}
