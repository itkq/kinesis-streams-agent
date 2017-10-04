package filewatcher

import (
	"log"
	"os"
	"syscall"
	"time"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/config"
	"github.com/itkq/kinesis-streams-agent/file_watcher/fswatcher"
	"github.com/itkq/kinesis-streams-agent/reader"
	"github.com/itkq/kinesis-streams-agent/reader/lifetimer"
	"github.com/itkq/kinesis-streams-agent/state"
)

type FileWatcher struct {
	config *config.FileWatcherConfig

	state state.State

	// fsnotify watcher
	fswatcher *fswatcher.Fswatcher

	// inode -> Reader
	readers map[uint64]reader.Reader

	// backup too big log entry
	backupIO *os.File

	// inode -> clockCh (connected to each reader)
	clockChMap map[uint64]chan<- time.Time

	// reader's output channel
	chunkCh chan<- *chunk.Chunk

	// file read interval
	ticker *time.Ticker

	newReaderFunc func(p string, i uint64, io *os.File, lt *lifetimer.LifeTimer) (reader.Reader, error)
}

func NewFileWatcher(
	conf *config.FileWatcherConfig,
	state state.State,
	chunkCh chan<- *chunk.Chunk,
) (*FileWatcher, error) {
	fswatcher, err := fswatcher.NewFswatcher()
	if err != nil {
		return nil, err
	}
	fswatcher.RegisterPaths(conf.WatchPaths)

	var backupIO *os.File
	if conf.UnputtableRecordsLocalBackupPath != "" {
		backupIO, err = os.OpenFile(
			conf.UnputtableRecordsLocalBackupPath,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC,
			0644,
		)
		if err != nil {
			return nil, err
		}
	}

	return &FileWatcher{
		config:        conf,
		state:         state,
		fswatcher:     fswatcher,
		readers:       make(map[uint64]reader.Reader),
		backupIO:      backupIO,
		clockChMap:    make(map[uint64]chan<- time.Time),
		chunkCh:       chunkCh,
		ticker:        time.NewTicker(conf.ReadFileInterval),
		newReaderFunc: reader.NewFileReader,
	}, nil
}

func (w *FileWatcher) Run(controlCh chan interface{}) {
	if err := w.InitReaders(); err != nil {
		log.Println("error:", err)
	}

	for {
		select {
		// file read clock
		case t := <-w.ticker.C:
			// propagate to existed each reader
			for inode, ch := range w.clockChMap {
				reader, ok := w.readers[inode]
				if ok && reader.Opened() {
					ch <- t
				} else {
					// deregister reader closed by itself already
					delete(w.readers, inode)
					delete(w.clockChMap, inode)
					log.Printf("info: reader %d deregisterd", inode)
				}
			}

		// filesystem event
		case ev := <-w.fswatcher.Events:
			if w.fswatcher.IsCreatedEvent(ev) && w.fswatcher.ShouldWatchEvent(ev) {
				path := ev.Name
				pinode := w.GetInode(path)
				if pinode == nil {
					log.Println("error: inode not found", path)
					continue
				}
				inode := *pinode

				// start watcher when watching file is created and reader not exist
				if _, ok := w.readers[inode]; !ok {
					w.StartReader(
						path,
						inode,
						w.newReaderFunc,
					)
				}
			}

		// filesystem error event
		case err := <-w.fswatcher.Errors:
			log.Println("error:", err)

		// return when signal received
		case <-controlCh:
			return
		}
	}
}

func (w *FileWatcher) InitReaders() error {
	for _, path := range w.fswatcher.ExpandPaths() {
		pinode := w.GetInode(path)
		if pinode == nil {
			log.Println("target file not found:", path)
			continue
		}
		err := w.StartReader(path, *pinode, w.newReaderFunc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *FileWatcher) StartReader(
	path string,
	inode uint64,
	newReaderFunc func(p string, i uint64, io *os.File, lt *lifetimer.LifeTimer) (reader.Reader, error),
) error {
	lifetimer := lifetimer.NewLifeTimer(path, inode)
	lifetimer.LifeTime = w.config.LifeTimeAfterMovedFile

	reader, err := newReaderFunc(path, inode, w.backupIO, lifetimer)
	if err != nil {
		return err
	}
	w.readers[inode] = reader

	clockCh := make(chan time.Time)
	w.clockChMap[inode] = clockCh

	readerState := w.state.GetReaderState(inode)
	if readerState == nil {
		readerState = w.state.CreateReaderState(inode, path)
	}

	go reader.Run(readerState, clockCh, w.chunkCh)
	log.Println("info: started reader", inode, path)

	return nil
}

func (w *FileWatcher) GetInode(path string) *uint64 {
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
