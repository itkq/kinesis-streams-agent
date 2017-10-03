kinesis-agent-go
====
Simple Amazon Kinesis Streams agent written in Go.

## Description
Collect logs based on `tail -F` and forward it to Amazon Kinesis Streams.

## Feature

### Tailing
Monitor filesystem event using [fsnotify](https://github.com/fsnotify/fsnotify) and
implement tailing like 
`tail -F` rather than `tail -f`.
Read position of each log is managed by a local file. 

### Aggregation
To reduce cost, log entries are aggregated (line-based) as one payload, [up to 25 KB](https://aws.amazon.com/kinesis/streams/pricing/).

### Fail safe
Read positions are updated only after logs are forwarded to Amazon Kinesis Streams successfully. If kinesis-agent-go has stopped unexpectedly, it forwards logs not sent yet when restarted.

## VS.

### [awslabs/amazon-kinesis-agent](https://github.com/awslabs/amazon-kinesis-agent)

Official agent but of course JVM is required.

## Requirement
- Linux (+ inotify)
- go1.8 or later

## Usage
```
$ kinesis-agent-go -c /path/to/config.yml
```

## Install
```
$ go get github.com/itkq/kinesis-agent-go
```

## Contribution

1. Fork ([https://github.com/itkq/kinesis-agent-go/fork](https://github.com/itkq/kinesis-agent-go/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `make test` command and confirm that it passes
6. Run `make fmt`
7. Create new Pull Request

## TODO
- improve sender's retry processing
- remove too old state entry 
- more tests
  - add end-to-end test case

## Licence

MIT

## Author

[itkq](https://github.com/itkq)
