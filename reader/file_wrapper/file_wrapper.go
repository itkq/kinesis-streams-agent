package file

import (
	"io"
	"os"
	"sync"
)

type FileWrapper struct {
	*os.File
	*sync.Mutex

	// The position relative to the origin of the file
	pos int64
}

func NewFileWrapper(f *os.File) *FileWrapper {
	return &FileWrapper{
		f,
		&sync.Mutex{},
		0,
	}
}

// SeekRel() is the wrapper of os.Seek()
// Seek to relative postion
func (fw *FileWrapper) SeekRel(offset int64) (ret int64, err error) {
	fw.pos += offset
	return fw.Seek(fw.pos, io.SeekStart)
}

// SeekAbs() is the wrapper of os.Seek()
// Seek to absolute postion
func (fw *FileWrapper) SeekAbs(pos int64) (ret int64, err error) {
	fw.pos = pos
	return fw.Seek(fw.pos, io.SeekStart)
}

// WriteBytes is the wrapper of os.Write()
func (fw *FileWrapper) WriteBytes(b []byte) (int, error) {
	_, err := fw.SeekRel(0)
	if err != nil {
		return 0, err
	}

	n, err := fw.Write(b)
	if err == nil {
		fw.pos += int64(n)
	}

	return n, err
}

func (fw *FileWrapper) WriteString(s string) (int, error) {
	return fw.WriteBytes([]byte(s))
}

func (fw *FileWrapper) ReadAtLeast(min int64) (int64, []byte, error) {
	buf := make([]byte, min)

	// ignore because err is only io.EOF
	n, _ := fw.Read(buf)

	if _, err := fw.SeekRel(int64(n)); err != nil {
		return int64(n), buf[:n], err
	}

	return int64(n), buf[:n], nil
}
