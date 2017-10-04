package local

import (
	"os"

	"github.com/itkq/kinesis-streams-agent/payload"
)

type LocalClient struct {
	io *os.File
}

func NewLocalClient(path string) (*LocalClient, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &LocalClient{
		io: f,
	}, nil
}

func (c *LocalClient) PutRecords(records []*payload.Record) ([]*payload.Record, error) {
	for _, r := range records {
		r.ErrorCode = nil
		c.io.Write(r.ToByte())
	}

	return records, nil
}
