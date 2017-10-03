package filewriter

import (
	"os"

	"github.com/itkq/kinesis-agent-go/payload"
)

type FileWriter struct {
	io *os.File
}

func NewFileWriter(path string) (*FileWriter, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		io: f,
	}, nil
}

func (f *FileWriter) PutRecords(records []*payload.Record) ([]*payload.Record, error) {
	for _, r := range records {
		r.ErrorCode = nil
		f.io.Write(r.ToByte())
	}

	return records, nil
}
