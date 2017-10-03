package sender

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/itkq/kinesis-agent-go/config"
	"github.com/itkq/kinesis-agent-go/payload"
	"github.com/itkq/kinesis-agent-go/sender/file_writer"
	"github.com/itkq/kinesis-agent-go/sender/kinesis"
	"github.com/itkq/kinesis-agent-go/sender/retry"
	"github.com/itkq/kinesis-agent-go/state"
)

const (
	DefaultRetryCountMax = 10
)

type SendClient interface {
	PutRecords(records []*payload.Record) ([]*payload.Record, error)
}

type Sender struct {
	client        SendClient
	state         state.State
	payloadCh     chan *payload.Payload
	sendInfosCh   chan []*state.SendInfo
	backoff       *retry.ExpBackOff
	RetryCountMax int
	retryRecords  []*payload.Record
}

func NewSender(
	conf *config.SenderConfig,
	state state.State,
	payloadCh chan *payload.Payload,
) (*Sender, error) {
	var err error
	var sendClient SendClient

	switch conf.Type {
	case "kinesis_streams":
		if conf.StreamName == "" {
			return nil, errors.New("stream_name must be set when kinesis_streams type")
		}

		awsConfig := aws.NewConfig()

		// configure forward proxy
		if conf.ForwardProxyUrl != "" {
			httpClient := &http.Client{
				Transport: &http.Transport{
					Proxy: func(*http.Request) (*url.URL, error) {
						return url.Parse(conf.ForwardProxyUrl)
					},
				},
			}
			awsConfig = awsConfig.WithHTTPClient(httpClient)
			log.Println("info: configured forward proxy: ", conf.ForwardProxyUrl)
		}

		ks, err := kinesis.NewKinesisStream(awsConfig)
		if err != nil {
			return nil, err
		}

		sendClient = kinesis.NewKinesisStreamClient(ks, &conf.StreamName)

	case "local":
		if conf.OutputFilePath == "" {
			return nil, errors.New("output_filepath must be set when local type")
		}

		sendClient, err = filewriter.NewFileWriter(conf.OutputFilePath)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("error: unknown type")
	}

	return &Sender{
		client:        sendClient,
		state:         state,
		payloadCh:     payloadCh,
		backoff:       retry.NewExpBackOff(),
		RetryCountMax: DefaultRetryCountMax,
		retryRecords:  make([]*payload.Record, 0),
	}, nil
}

func (s *Sender) Run() {
	for {
		p := <-s.payloadCh
		err := s.SendWithRetry(p.Records)
		if err != nil {
			log.Println("error:", err)
			os.Exit(1)
		}
	}
}

func (s *Sender) SendWithRetry(records []*payload.Record) error {
	retryRecords := records

	s.backoff.Reset()
	return retry.Retry(s.RetryCountMax, s.backoff, func() error {
		resultRecords := s.Send(retryRecords)
		retryRecords = make([]*payload.Record, 0)
		for i, _ := range resultRecords {
			r := *resultRecords[i]
			if r.ErrorCode != (*string)(nil) {
				retryRecords = append(retryRecords, &r)
			}
		}
		s.retryRecords = retryRecords
		if len(s.retryRecords) == 0 {
			s.retryRecords = make([]*payload.Record, 0)
			return nil
		}
		return fmt.Errorf("retry")
	})
}

// returns failed records
func (s *Sender) Send(records []*payload.Record) []*payload.Record {
	responseRecords, err := s.client.PutRecords(records)
	if err != nil {
		log.Println("error:", err)
	}

	for _, r := range responseRecords {
		if r.ErrorCode == (*string)(nil) {
			r.Success()
		}
		for _, c := range r.Chunks {
			s.state.Update(c.SendInfo)
		}
	}

	if err := s.state.DumpToJSON(); err != nil {
		log.Println("error:", err)
	}

	return responseRecords
}
