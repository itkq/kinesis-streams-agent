package filewriter

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/payload"
	"github.com/itkq/kinesis-agent-go/state"
	"github.com/stretchr/testify/assert"
)

func TestPutRecords(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "sender")
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "test_output")
	writer, err := NewFileWriter(outputPath)
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}

	content := []string{
		"hoge\n",
		"fuga\n",
		"piyo\n",
		"poyo\n",
	}

	records := []*payload.Record{
		&payload.Record{
			Size: 15,
			Chunks: []*chunk.Chunk{
				&chunk.Chunk{
					SendInfo: &state.SendInfo{
						ReadRange: &state.FileReadRange{},
					},
					Body: []byte(strings.Join(content[0:2], "")),
				},
				&chunk.Chunk{
					SendInfo: &state.SendInfo{
						ReadRange: &state.FileReadRange{},
					},
					Body: []byte(content[2]),
				},
			},
		},
		&payload.Record{
			Chunks: []*chunk.Chunk{
				&chunk.Chunk{
					SendInfo: &state.SendInfo{
						ReadRange: &state.FileReadRange{},
					},
					Body: []byte(content[3]),
				},
			},
		},
	}

	responseRecords, err := writer.PutRecords(records)
	assert.NoError(t, err)
	for _, r := range responseRecords {
		assert.Nil(t, r.ErrorCode)
	}
	b, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	assert.Equal(t, []byte(strings.Join(content, "")), b)
}
