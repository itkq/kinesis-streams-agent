package kinesis

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/itkq/kinesis-agent-go/payload"
	uuid "github.com/satori/go.uuid"
)

const (
	RecordCountMax = 500

	// 1 MB
	RecordSizeMax = 1 * 1024 * 1024

	// 5 MB
	EntireRequestSizeMax = 5 * 1024 * 1024

	// 25 KB
	PutPayloadUnitSize = 25 * 1024
)

var ErrorCodes = map[string]struct{}{
	"ProvisionedThroughputExceededException": {},
	"InternalFailure":                        {},
}

func NewPutRecordsRequestEntry(
	// required: blob
	data []byte,
	// required: partitionKey is Unicode string, with a maximum length limit of
	// 256 characters
	partitionKey *string,
	// optional: used for overriding partitonKey
	explicitHashKey *string,
) *kinesis.PutRecordsRequestEntry {
	return &kinesis.PutRecordsRequestEntry{
		Data:            data,
		PartitionKey:    partitionKey,
		ExplicitHashKey: explicitHashKey,
	}
}

type KinesisStreamsClient struct {
	kinesis    KinesisStreamsClientIface
	StreamName *string
}

type KinesisStreamsClientIface interface {
	PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
}

func NewKinesisStream(c *aws.Config) (*kinesis.Kinesis, error) {
	sess, err := session.NewSession(c)
	if err != nil {
		return nil, err
	}

	return kinesis.New(sess), nil
}

func NewKinesisStreamClient(
	kinesis KinesisStreamsClientIface,
	streamName *string,
) *KinesisStreamsClient {
	return &KinesisStreamsClient{
		kinesis:    kinesis,
		StreamName: streamName,
	}
}

func (k *KinesisStreamsClient) PutRecords(records []*payload.Record) ([]*payload.Record, error) {
	requestEntriesMap := make(map[*payload.Record]*string)
	requestEntries := make([]*kinesis.PutRecordsRequestEntry, 0, len(records))

	for _, r := range records {
		// UUID is used as the partition key
		partitonKey := uuid.NewV4().String()

		entry := NewPutRecordsRequestEntry(r.ToByte(), &partitonKey, nil)
		requestEntries = append(requestEntries, entry)
		requestEntriesMap[r] = &partitonKey
	}

	resultEntries, err := k.putRecords(requestEntries)

	retRecords := make([]*payload.Record, len(records))
	for i, r := range resultEntries {
		retRecords[i] = records[i]
		retRecords[i].ErrorCode = r.ErrorCode
		retRecords[i].ErrorMessage = r.ErrorMessage
	}

	return retRecords, err
}

func (k *KinesisStreamsClient) putRecords(
	records []*kinesis.PutRecordsRequestEntry,
) ([]*kinesis.PutRecordsResultEntry, error) {
	input := &kinesis.PutRecordsInput{
		Records:    records,
		StreamName: k.StreamName,
	}

	// aws-sdk-go retries with exponential backoff by default
	output, err := k.kinesis.PutRecords(input)
	return output.Records, err
}
