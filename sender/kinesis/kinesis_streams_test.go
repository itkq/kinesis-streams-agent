package kinesis

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/payload"
	"github.com/itkq/kinesis-streams-agent/state"
	"github.com/stretchr/testify/assert"
)

type fakeKinesisStreams struct {
	KinesisStreamsClientIface
	FakePutRecords func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
}

func (c *fakeKinesisStreams) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	return c.FakePutRecords(input)
}

var records []*payload.Record = []*payload.Record{
	&payload.Record{
		Size: 5,
		Chunks: []*chunk.Chunk{
			&chunk.Chunk{
				SendInfo: &state.SendInfo{
					ReadRange: &state.FileReadRange{
						Begin: 0,
						End:   5,
					},
				},
				Body: []byte("hoge\n"),
			},
		},
	},
	&payload.Record{
		Size: 5,
		Chunks: []*chunk.Chunk{
			&chunk.Chunk{
				SendInfo: &state.SendInfo{
					ReadRange: &state.FileReadRange{
						Begin: 5,
						End:   10,
					},
				},
				Body: []byte("fuga\n"),
			},
		},
	},
	&payload.Record{
		Size: 5,
		Chunks: []*chunk.Chunk{
			&chunk.Chunk{
				SendInfo: &state.SendInfo{
					ReadRange: &state.FileReadRange{
						Begin: 10,
						End:   15,
					},
				},
				Body: []byte("piyo\n"),
			},
		},
	},
}

func TestPutRecordsWithNoError(t *testing.T) {
	fakeKinesis := &fakeKinesisStreams{
		FakePutRecords: func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
			entries := make([]*kinesis.PutRecordsResultEntry, len(input.Records))
			for i, _ := range input.Records {
				entries[i] = &kinesis.PutRecordsResultEntry{
					ErrorCode:    nil,
					ErrorMessage: nil,
				}
			}
			return &kinesis.PutRecordsOutput{
				FailedRecordCount: &[]int64{0}[0],
				Records:           entries,
			}, nil
		},
	}

	client := NewKinesisStreamClient(fakeKinesis, nil)
	resultRecords, err := client.PutRecords(records)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(records), len(resultRecords))
	for _, r := range resultRecords {
		assert.Nil(t, r.ErrorCode)
		assert.Nil(t, r.ErrorMessage)
	}
}

func TestPutRecordWithOneFailedRecord(t *testing.T) {
	rand.Seed(time.Now().Unix())
	fakeKinesis := &fakeKinesisStreams{
		FakePutRecords: func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
			dummyString := "dummy"
			size := len(input.Records)
			failedIndex := rand.Intn(size)

			entries := make([]*kinesis.PutRecordsResultEntry, len(input.Records))
			for i, _ := range input.Records {
				if i == failedIndex {
					entries[i] = &kinesis.PutRecordsResultEntry{
						ErrorCode:    &dummyString,
						ErrorMessage: &dummyString,
					}
				} else {
					entries[i] = &kinesis.PutRecordsResultEntry{
						ErrorCode:    nil,
						ErrorMessage: nil,
					}
				}
			}

			return &kinesis.PutRecordsOutput{
				FailedRecordCount: &[]int64{1}[0],
				Records:           entries,
			}, nil
		},
	}

	client := NewKinesisStreamClient(fakeKinesis, nil)
	resultRecords, err := client.PutRecords(records)
	assert.Equal(t, nil, err)

	successCnt := 0
	for _, r := range resultRecords {
		if r.ErrorCode == (*string)(nil) {
			successCnt++
		}
	}
	assert.Equal(t, len(records)-1, successCnt)
}

func TestPutRecordWithKinesisStreamsError(t *testing.T) {
	fakeKinesis := &fakeKinesisStreams{
		FakePutRecords: func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
			dummyString := "dummy"
			entries := make([]*kinesis.PutRecordsResultEntry, len(input.Records))
			for i, _ := range input.Records {
				entries[i] = &kinesis.PutRecordsResultEntry{
					ErrorCode:    &dummyString,
					ErrorMessage: &dummyString,
				}
			}

			cnt := int64(len(entries))
			return &kinesis.PutRecordsOutput{
				FailedRecordCount: &cnt,
				Records:           entries,
			}, errors.New("dummy")
		},
	}

	client := NewKinesisStreamClient(fakeKinesis, nil)

	resultRecords, err := client.PutRecords(records)
	assert.NotEqual(t, nil, err)
	for _, r := range resultRecords {
		assert.NotNil(t, r.ErrorCode)
	}
}
