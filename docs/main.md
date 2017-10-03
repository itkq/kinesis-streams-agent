# Components

Components are struct implemented `Run()` and called as goroutine.

## FileWatcher

- Summary
  - Watch target files using fsnotify and manage FileReaders.

- Detail
  - If new target file is found, new FileReader will be started and registered.
  - Notify to opened FileReaders through each channel at a specific interval. Closed FileReaders are deregistered.

## FileReader
- Summary
  - Read logs and send _chunk_ to Aggregator.
 
- Detail
  - When FileReader received a notification from FileWatcher, FileReader try to read new logs (line-based). If new logs are found, FileReader convert these logs to _chunk_  (struct, log byte and read range are included) and send it to Aggregator through a channel.
  - When the file is removed or rotated, a lifetimer is started. If a configured lifetime is elapsed and there are no new logs, FileReader will be closed.

## Aggregator
- Summary
  - Aggregate _chunks_ to _payload_ and send _payload_ to Sender.

- Detail
  - Received _chunk_ is stored into _PayloadBuffer_.
  - _chunks_ are stored into one _record_ (struct, set of _chunk_) up to 25 KB.
  - _records_ are stored into one _payload_ (struct, set of _record_) up to 5 MB or 500 _records_.
  - When specific interval is elapsed or current _payload_ is full, it is sent to Sender through a channel.

## Sender
- Summary
  - Send _records_ to Kinesis Streams.

- Detail
  - Received _records_ are sent using `PutRecords` by KinesisStreamsClient.
  - Update _state_ with both succeeded _records_ and failed _records_ 

# State
If sending logs to Kinesis Streams is failed, these logs should be retry to send.
These logs are lost when the agent is terminated abnormaly while retrying to send it.
For this reason, it is better to manage the information of sent logs in local file.
State file is equivalent of that file.

Sample:
```
$ cat test.state
{
    "8719120": {
        "pos": 14,
        "path": "/tmp/kinesis-agent-go/test1.log",
        "send_ranges": [
            {
                "begin": 0,
                "end": 14
            }
        ]
    },
    "8719122": {
        "pos": 52,
        "path": "/tmp/kinesis-agent-go/test2.log",
        "send_ranges": [
            {
                "begin": 0,
                "end": 28
            }
        ]
    }
}
```

# Sequence

## Main flow

![](https://github.com/itkq/kinesis-agent-go/raw/master/docs/plantuml/main.png)
