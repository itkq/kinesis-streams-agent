package filewatcher

import (
	"github.com/itkq/kinesis-agent-go/reader"
)

func (w *FileWatcher) Endpoint() string {
	return "/file_watcher"
}

func (w *FileWatcher) Export() interface{} {
	readers := make(map[uint64]interface{})
	for i, r := range w.readers {
		readers[i] = r.(*reader.FileReader).Export()
	}

	return &FileWatcherMetrics{
		Readers: readers,
	}
}

type FileWatcherMetrics struct {
	Readers map[uint64]interface{} `json:"readers"`
}
