# kinesis-streams-agent

Amazon Kinesis Streams agent written in Go.

[![CircleCI](https://circleci.com/gh/itkq/kinesis-streams-agent.svg?style=svg)](https://circleci.com/gh/itkq/kinesis-streams-agent)

## Description
Collect logs based on `tail -F` and send it to Amazon Kinesis Streams.

## Feature

### Tailing
Monitor filesystem event using [fsnotify](https://github.com/fsnotify/fsnotify) and
implement tailing like
`tail -F` rather than `tail -f`.
Read position of each log is managed by a local file.

### Aggregation
To reduce cost, log entries are aggregated (line-based) as one record, [up to 25 KB](https://aws.amazon.com/kinesis/streams/pricing/).

### At Lest Once
Sent positions are updated immediately after logs are sent to Amazon Kinesis Streams using PutRecords API.
Failed records are saved on-memory and retried to send by exponential backoff.
If kinesis-streams-agent has stopped unexpectedly, it send logs not sent yet when restarted.

## VS.

### [awslabs/amazon-kinesis-agent](https://github.com/awslabs/amazon-kinesis-agent)

Official agent but of course JVM is required.

## Requirement
- Linux (+ inotify)
- go1.8 or later

## Usage
```
$ kinesis-streams-agent -c /path/to/config.yml
```

## Install
```
$ go get github.com/itkq/kinesis-streams-agent
```

## Contribution

1. Fork ([https://github.com/itkq/kinesis-streams-agent/fork](https://github.com/itkq/kinesis-streams-agent/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `make test` command and confirm that it passes
6. Run `make fmt`
7. Create new Pull Request

## TODO
- Add end-to-end test case

## Licence

MIT

## Author

[itkq](https://github.com/itkq)
