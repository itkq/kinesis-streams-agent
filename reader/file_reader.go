package reader

import (
	"errors"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/itkq/kinesis-streams-agent/chunk"
	"github.com/itkq/kinesis-streams-agent/reader/file_wrapper"
	"github.com/itkq/kinesis-streams-agent/reader/lifetimer"
	"github.com/itkq/kinesis-streams-agent/sender/kinesis"
	"github.com/itkq/kinesis-streams-agent/state"
)

const (
	NewLineRune        = '\n'
	ReadByteSize       = 2048
	FileOpenPermission = 0644
	DefaultMaxLineSize = kinesis.RecordSizeMax
)

type FileReader struct {
	path        string
	inode       uint64
	pos         int64
	io          *file.FileWrapper
	chunkCh     chan<- *chunk.Chunk // output channel
	backupIO    *os.File
	lifetimer   *lifetimer.LifeTimer
	MaxLineSize int64
}

func NewFileReader(
	path string,
	inode uint64,
	backupIO *os.File,
	lifetimer *lifetimer.LifeTimer,
) (Reader, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, FileOpenPermission)
	if err != nil {
		return nil, err
	}

	io := file.NewFileWrapper(f)

	w := &FileReader{
		path:        path,
		inode:       inode,
		io:          io,
		backupIO:    backupIO,
		lifetimer:   lifetimer,
		MaxLineSize: DefaultMaxLineSize,
	}

	return w, nil
}

func (r *FileReader) Run(
	readerState *state.ReaderState,
	// input channel
	clockCh <-chan time.Time,
	// output channel
	chunkCh chan<- *chunk.Chunk,
) {
	r.chunkCh = chunkCh
	if err := r.InitialRead(readerState); err != nil {
		log.Println("error:", err)
		return
	}

	r.pos = readerState.Pos
	for {
		_, ok := <-clockCh
		if !ok {
			log.Printf("reader (%s, %d) closed", r.path, r.inode)
			return
		}

		chunk, err := r.ReadLines()

		// lifetimer starts when file removed or file rotated.
		// reader closes when lifetime_after_file_moved elapsed
		// and there are no new line.
		if err != nil || r.Rotated() {
			if r.lifetimer.ShouldDie() && chunk == nil {
				r.Close()
				return
			}
		}

		if chunk != nil {
			r.chunkCh <- chunk
		}
	}
}

func (r *FileReader) InitialRead(rstate *state.ReaderState) error {
	for _, readRange := range rstate.LeakedRanges() {
		chunk, err := r.ReadLinesInRange(readRange)
		if err != nil {
			return err
		} else {
			r.chunkCh <- chunk
		}
	}

	return nil
}

func (r *FileReader) ReadLines() (*chunk.Chunk, error) {
	n, bytes, err := r.readBytesByLine(r.pos)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	readRange := &state.FileReadRange{
		Begin: r.pos,
		End:   r.pos + n,
	}

	r.pos += n

	return &chunk.Chunk{
		SendInfo: &state.SendInfo{
			Inode:     r.inode,
			ReadRange: readRange,
		},
		Body: bytes,
	}, nil
}

func (r *FileReader) ReadLinesInRange(readRange *state.FileReadRange) (*chunk.Chunk, error) {
	_, b, err := r.readBytesByLineInRange(readRange.Begin, readRange.End)
	if err != nil {
		return nil, err
	}

	return &chunk.Chunk{
		SendInfo: &state.SendInfo{
			Inode:     r.inode,
			ReadRange: readRange,
		},
		Body: b,
	}, nil
}

func (r *FileReader) readBytesByLineInRange(start int64, end int64) (int64, []byte, error) {
	// check file exists
	if _, err := os.Stat(r.path); err != nil {
		return 0, []byte{}, err
	}

	size := end - start
	r.io.SeekAbs(start)

	n, b, err := r.io.ReadAtLeast(size)
	if err != nil {
		return n, b, err
	}
	if n != size {
		return n, b, errors.New("ReadLinesInRange read size mismatch")
	}

	return n, b, err
}

func (r *FileReader) readBytesByLine(start int64) (int64, []byte, error) {
	var buf []byte = make([]byte, 0)
	var lastNewLinePos int64 = -1
	var readSize int64

	// check file exists
	if _, err := os.Stat(r.path); err != nil {
		return 0, buf, err
	}

	r.io.SeekAbs(start)

	for readSize = 0; ; {
		n, b, err := r.io.ReadAtLeast(ReadByteSize)
		if err != nil {
			return readSize, buf, err
		}

		// EOF
		if n == 0 {
			return readSize, buf, err
		}

		for i := 0; i < len(b); i++ {
			if b[i] == NewLineRune {
				line := append(buf[lastNewLinePos+1:len(buf)], b[0:i+1]...)
				lineSize := int64(len(line))

				if lineSize > r.MaxLineSize {
					log.Printf(
						"line size is over %d: %s",
						r.MaxLineSize,
						line,
					)
					if r.backupIO != nil {
						r.backupIO.Write(line)
					}

					buf = buf[0 : lastNewLinePos+1]
					b = b[i+1:]
					i = 0
				} else {
					buf = append(buf, b[0:i+1]...)
					readSize += lineSize
					lastNewLinePos += lineSize

					b = b[i+1:]
					i = 0
				}
			}
		}

		if n < ReadByteSize {
			break
		}

		if len(b) > 0 {
			buf = append(buf, b...)
		}
	}

	return readSize, buf, nil
}

func (r *FileReader) Rotated() bool {
	info, err := os.Stat(r.path)
	if err != nil {
		return true
	}
	stat, _ := info.Sys().(*syscall.Stat_t)
	if r.inode != stat.Ino {
		return true
	}

	return false
}

func (r *FileReader) Opened() bool {
	return r.io != nil
}

func (r *FileReader) Close() {
	r.io.Close()
	r.io = nil
}

func (r *FileReader) Export() interface{} {
	return &FileReaderMetrics{
		Pos:  r.pos,
		Path: r.path,
	}
}

type FileReaderMetrics struct {
	Pos  int64  `json:"pos"`
	Path string `json:"path"`
}
