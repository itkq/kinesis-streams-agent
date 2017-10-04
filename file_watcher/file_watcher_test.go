package filewatcher

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/itkq/kinesis-streams-agent/reader/lifetimer"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/config"
	"github.com/itkq/kinesis-streams-agent/reader"
	"github.com/itkq/kinesis-streams-agent/state"
	"github.com/stretchr/testify/assert"
)

var lifeTimeAfterFileMoved time.Duration = 500 * time.Millisecond

var configTemplate config.FileWatcherConfig = config.FileWatcherConfig{
	ReadFileInterval: 200 * time.Millisecond,
}

type dummyReader struct {
	path   string
	inode  uint64
	opened bool
}

func newDummyReader(p string, i uint64, io *os.File, lt *lifetimer.LifeTimer) (reader.Reader, error) {
	return &dummyReader{
		path:   p,
		inode:  i,
		opened: true,
	}, nil
}

func (r *dummyReader) Run(
	readerState *state.ReaderState,
	clockCh <-chan time.Time,
	chunkCh chan<- *chunk.Chunk,
) {
	timerCh := time.After(lifeTimeAfterFileMoved)
	for {
		select {
		case <-timerCh:
			r.opened = false
			return

		case <-clockCh:
		}
	}
}

func (r *dummyReader) Opened() bool {
	return r.opened
}

func TestRun(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_watcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	conf := configTemplate
	watchPath := filepath.Join(dir, "test*.log")
	conf.WatchPaths = []string{
		watchPath,
	}

	chunkCh := make(chan<- *chunk.Chunk)
	watcher, err := NewFileWatcher(&conf, &state.DummyState{}, chunkCh)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	watcher.newReaderFunc = newDummyReader

	controlCh := make(chan interface{})

	go watcher.Run(controlCh)

	fn1 := filepath.Join(dir, "test1.log")
	f1, err := os.OpenFile(fn1, os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	f1.Close()

	time.Sleep(200 * time.Millisecond)

	i1 := *watcher.GetInode(fn1)
	r1, ok := watcher.readers[i1]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, r1.Opened())

	time.Sleep(lifeTimeAfterFileMoved * 2)
	_, ok = watcher.readers[i1]
	assert.Equal(t, false, ok)

	fn2 := filepath.Join(dir, "test2.log")
	f2, err := os.OpenFile(fn2, os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	f2.Close()
	i2 := *watcher.GetInode(fn2)

	time.Sleep(200 * time.Millisecond)

	r2, ok := watcher.readers[i2]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, r2.Opened())

	time.Sleep(lifeTimeAfterFileMoved * 2)
	_, ok = watcher.readers[i1]
	assert.Equal(t, false, ok)
}

func TestInitReaders(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_watcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	conf := configTemplate
	watchPath := filepath.Join(dir, "test*.log")
	conf.WatchPaths = []string{
		watchPath,
	}

	chunkCh := make(chan *chunk.Chunk)
	watcher, err := NewFileWatcher(&conf, &state.DummyState{}, chunkCh)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	watcher.newReaderFunc = newDummyReader

	fn := filepath.Join(dir, "test1.log")
	f, err := os.OpenFile(fn, os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	f.Close()

	pinode := watcher.GetInode(fn)
	watcher.InitReaders()

	inode := *pinode
	_, ok := watcher.clockChMap[inode]
	assert.Equal(t, true, ok)
	reader, ok := watcher.readers[inode]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, reader.Opened())
}

func TestStartWatcher(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_watcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	conf := configTemplate
	watchPath := filepath.Join(dir, "test*.log")
	conf.WatchPaths = []string{
		watchPath,
	}

	chunkCh := make(chan *chunk.Chunk)
	watcher, err := NewFileWatcher(&conf, &state.DummyState{}, chunkCh)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fn := filepath.Join(dir, "test1.log")
	f, err := os.OpenFile(fn, os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	f.Close()

	inode := getInode(fn)
	err = watcher.StartReader(fn, *inode, newDummyReader)
	assert.Equal(t, nil, err)
	_, ok := watcher.clockChMap[*inode]
	assert.Equal(t, true, ok)
	reader, ok := watcher.readers[*inode]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, reader.Opened())
}

func getInode(path string) *uint64 {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	var ok bool
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}

	return &stat.Ino
}
