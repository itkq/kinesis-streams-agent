package sender

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/itkq/kinesis-agent-go/chunk"
	"github.com/itkq/kinesis-agent-go/payload"
	"github.com/itkq/kinesis-agent-go/sender/local"
	"github.com/itkq/kinesis-agent-go/sender/retry"
	"github.com/itkq/kinesis-agent-go/state"
	"github.com/stretchr/testify/assert"
)

var tmpDir string
var sender *Sender
var exitCode = 0

type SendOnlyOneRecordClient struct{}

func (c *SendOnlyOneRecordClient) PutRecords(
	records []*payload.Record,
) ([]*payload.Record, error) {
	rand.Seed(time.Now().Unix())

	size := len(records)
	if size == 0 {
		return records, nil
	}

	successIndex := rand.Intn(size)
	dummyString := "dummy"

	retRecords := records
	for i, r := range retRecords {
		if i != successIndex {
			r.ErrorCode = &dummyString
			r.ErrorMessage = &dummyString
		} else {
			r.ErrorCode = nil
			r.ErrorMessage = nil
		}
	}

	return retRecords, fmt.Errorf(dummyString)
}

func TestRun(t *testing.T) {
	dir, err := ioutil.TempDir("", "sender")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	outputPath := filepath.Join(dir, "test_output")
	client, err := local.NewLocalClient(outputPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	sender := NewSender(client, &state.DummyState{}, make(chan *payload.Payload))

	go sender.Run()

	sender.payloadCh <- &payload.Payload{
		Size:  0,
		Count: 2,
		Records: []*payload.Record{
			&payload.Record{
				Chunks: []*chunk.Chunk{
					&chunk.Chunk{
						SendInfo: &state.SendInfo{},
						Body:     []byte("hoge\n"),
					},
				},
			},
			&payload.Record{
				Chunks: []*chunk.Chunk{
					&chunk.Chunk{
						SendInfo: &state.SendInfo{},
						Body:     []byte("fuga\n"),
					},
				},
			},
		},
	}

	b, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	assert.Equal(t, []byte("hoge\nfuga\n"), b)
}

func TestSendWithRetry(t *testing.T) {
	sender := &Sender{
		client:        &SendOnlyOneRecordClient{},
		state:         &state.DummyState{},
		payloadCh:     make(chan *payload.Payload),
		backoff:       retry.NewExpBackOff(),
		RetryCountMax: 3,
	}

	records := []*payload.Record{
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
	}

	err := sender.SendWithRetry(records)
	assert.NoError(t, err)
	assert.Equal(t, len(records)-1, sender.backoff.GetRetryCount())

	records = []*payload.Record{
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
		payload.NewRecord(),
	}
	err = sender.SendWithRetry(records)
	assert.Error(t, err)
	assert.Equal(t, 1, len(sender.retryRecords))
}

func TestSend(t *testing.T) {
	sender := &Sender{
		client:    &SendOnlyOneRecordClient{},
		state:     &state.DummyState{},
		payloadCh: make(chan *payload.Payload),
		backoff:   retry.NewExpBackOff(),
	}

	records := []*payload.Record{
		payload.NewRecord(),
		payload.NewRecord(),
	}

	records = sender.Send(records)
	assert.Equal(t, 2, len(records))

	// TODO: sendinfo check

	records = extractFailedRecords(records)
	assert.Equal(t, 1, len(records))

	records = sender.Send(records)
	assert.Equal(t, 1, len(records))
	records = extractFailedRecords(records)
	assert.Equal(t, 0, len(records))

	records = sender.Send(records)
	assert.Equal(t, 0, len(records))
}

func extractFailedRecords(records []*payload.Record) []*payload.Record {
	ret := make([]*payload.Record, 0)
	for i, _ := range records {
		r := *records[i]
		if r.ErrorCode != (*string)(nil) {
			ret = append(ret, &r)
		}
	}

	return ret
}
