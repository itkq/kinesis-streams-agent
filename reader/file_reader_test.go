package reader

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/itkq/kinesis-agent-go/chunk"
	file "github.com/itkq/kinesis-agent-go/reader/file_wrapper"
	"github.com/itkq/kinesis-agent-go/reader/lifetimer"
	"github.com/itkq/kinesis-agent-go/state"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_reader")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	checkErr(err)

	fn := filepath.Join(dir, "test1.log")
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	checkErr(err)

	pinode := getInode(fn)
	lt := lifetimer.NewLifeTimer(fn, *pinode)
	lt.LifeTime = 100 * time.Millisecond
	reader, err := NewFileReader(fn, *pinode, nil, lt)

	content1 := "hoge\n"
	f.WriteString(content1)

	readerState := &state.ReaderState{
		Pos:        0,
		SendRanges: []*state.FileReadRange{},
	}
	clockCh := time.NewTicker(200 * time.Millisecond).C
	chunkCh := make(chan *chunk.Chunk)
	go reader.Run(readerState, clockCh, chunkCh)

	assert.True(t, reader.Opened())

	var chunk *chunk.Chunk
	chunk = <-chunkCh
	assert.Equal(t, content1, string(chunk.Body))

	content2 := "fuga\n"
	f.WriteString(content2)
	content3 := "piyo\n"
	f.WriteString(content3)

	chunk = <-chunkCh
	assert.Equal(t, content2+content3, string(chunk.Body))
	os.Remove(fn)
	assert.True(t, reader.Opened())

	time.Sleep(500 * time.Millisecond)
	assert.False(t, reader.Opened())
}

func TestInitialRead(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_reader")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test1.log")
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	pinode := getInode(fn)
	reader := newFileReader(fn, *pinode)

	ranges := []*state.FileReadRange{
		&state.FileReadRange{
			Begin: 2,
			End:   4,
		},
		&state.FileReadRange{
			Begin: 6,
			End:   8,
		},
	}

	content := "01234567890"

	readerState := &state.ReaderState{
		Pos:        int64(len(content) - 1),
		SendRanges: ranges,
	}

	f.WriteString(content)

	chunkCh := make(chan *chunk.Chunk)
	reader.chunkCh = chunkCh
	go reader.InitialRead(readerState)

	var chunk *chunk.Chunk

	chunk = <-chunkCh
	assert.Equal(
		t, &state.FileReadRange{Begin: 0, End: 2}, chunk.SendInfo.ReadRange,
	)
	assert.Equal(t, []byte(content[0:2]), chunk.Body)

	chunk = <-chunkCh
	assert.Equal(
		t, &state.FileReadRange{Begin: 4, End: 6}, chunk.SendInfo.ReadRange,
	)
	assert.Equal(t, []byte(content[4:6]), chunk.Body)

	chunk = <-chunkCh
	assert.Equal(
		t, &state.FileReadRange{Begin: 8, End: 10}, chunk.SendInfo.ReadRange,
	)
	assert.Equal(t, []byte(content[8:10]), chunk.Body)
}

func TestReadInRange(t *testing.T) {
	type ReadInRangeTestCase struct {
		content       string
		start         int64
		end           int64
		expectedN     int64
		expectedBytes []byte
		expectedError error
		desc          string
	}

	var testCases = []*ReadInRangeTestCase{
		&ReadInRangeTestCase{
			content:       "hoge\nfuga\npiyo\n",
			start:         5,
			end:           10,
			expectedN:     5,
			expectedBytes: []byte("fuga\n"),
			expectedError: nil,
			desc:          "test intermediate",
		},
		&ReadInRangeTestCase{
			content:       "hoge\nfuga\npiyo\n",
			start:         0,
			end:           10,
			expectedN:     10,
			expectedBytes: []byte("hoge\nfuga\n"),
			expectedError: nil,
			desc:          "test beginning of line",
		},
		&ReadInRangeTestCase{
			content:       "hoge\nfuga\npiyo\n",
			start:         5,
			end:           15,
			expectedN:     10,
			expectedBytes: []byte("fuga\npiyo\n"),
			expectedError: nil,
			desc:          "test end of line",
		},
	}

	dir, err := ioutil.TempDir("", "file_reader")
	checkErr(err)
	defer os.RemoveAll(dir)

	var reader *FileReader
	for i, c := range testCases {
		fn := filepath.Join(dir, fmt.Sprintf("test%d.log", i))
		f, err := os.OpenFile(
			fn,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC,
			FileOpenPermission,
		)
		checkErr(err)

		pinode := getInode(fn)
		reader = newFileReader(fn, *pinode)

		f.WriteString(c.content)
		n, bytes, err := reader.readBytesByLineInRange(c.start, c.end)
		assert.Equal(t, c.expectedN, n, fmt.Sprintf("%s: test size", c.desc))
		assert.Equal(t, c.expectedBytes, bytes, fmt.Sprintf("%s: test lines", c.desc))
		assert.Equal(t, c.expectedError, err, fmt.Sprintf("%s: test error", c.desc))
	}
}

func TestReadBytesByLine(t *testing.T) {
	type ReadBytesByLineTestCase struct {
		content       string
		start         int64
		maxLineSize   int64
		expectedN     int64
		expectedBytes []byte
		expectedError error
		desc          string
	}

	var testCases = [][]*ReadBytesByLineTestCase{
		[]*ReadBytesByLineTestCase{
			&ReadBytesByLineTestCase{
				content:       "hoge\n",
				start:         0,
				maxLineSize:   4096,
				expectedN:     5,
				expectedBytes: []byte("hoge\n"),
				expectedError: nil,
				desc:          "simple test: read one line",
			},
			&ReadBytesByLineTestCase{
				content:       "hoge\nfuga\n",
				start:         5,
				maxLineSize:   4096,
				expectedN:     10,
				expectedBytes: []byte("hoge\nfuga\n"),
				expectedError: nil,
				desc:          "simple test: read two lines",
			},
			&ReadBytesByLineTestCase{
				content:       "hoge\nfuga",
				start:         15,
				maxLineSize:   4096,
				expectedN:     5,
				expectedBytes: []byte("hoge\n"),
				expectedError: nil,
				desc:          "simple test: read one line skipping next line with no newline",
			},
			&ReadBytesByLineTestCase{
				content:       "",
				start:         20,
				maxLineSize:   4096,
				expectedN:     0,
				expectedBytes: []byte{},
				expectedError: nil,
				desc:          "simple test: read no lines",
			},
		},
		[]*ReadBytesByLineTestCase{
			&ReadBytesByLineTestCase{
				content:       strings.Repeat("testfoo\n", 255),
				start:         0,
				maxLineSize:   4096,
				expectedN:     8 * 255,
				expectedBytes: []byte(strings.Repeat("testfoo\n", 255)),
				expectedError: nil,
				desc:          "bound test1: read all lines",
			},
			&ReadBytesByLineTestCase{
				content:       "",
				start:         8 * 255,
				maxLineSize:   4096,
				expectedN:     0,
				expectedBytes: []byte{},
				expectedError: nil,
				desc:          "bound test1: read no lines",
			},
		},
		[]*ReadBytesByLineTestCase{
			&ReadBytesByLineTestCase{
				content:       strings.Repeat("testfoo\n", 256),
				start:         0,
				maxLineSize:   4096,
				expectedN:     8 * 256,
				expectedBytes: []byte(strings.Repeat("testfoo\n", 256)),
				expectedError: nil,
				desc:          "bound test2: read all lines",
			},
			&ReadBytesByLineTestCase{
				content:       "",
				start:         8 * 256,
				maxLineSize:   4096,
				expectedN:     0,
				expectedBytes: []byte{},
				expectedError: nil,
				desc:          "bound test2: read no lines",
			},
		},
		[]*ReadBytesByLineTestCase{
			&ReadBytesByLineTestCase{
				content:       strings.Repeat("testfoo\n", 257),
				start:         0,
				maxLineSize:   4096,
				expectedN:     8 * 257,
				expectedBytes: []byte(strings.Repeat("testfoo\n", 257)),
				expectedError: nil,
				desc:          "bound test3: read all lines",
			},
			&ReadBytesByLineTestCase{
				content:       "",
				start:         8 * 257,
				maxLineSize:   4096,
				expectedN:     0,
				expectedBytes: []byte{},
				expectedError: nil,
				desc:          "bound test3: read no lines",
			},
		},
		[]*ReadBytesByLineTestCase{
			&ReadBytesByLineTestCase{
				content:       "hoge\n" + strings.Repeat("a", 1024*1024) + "\nfuga\n",
				start:         0,
				maxLineSize:   1024 * 1024,
				expectedN:     5 + 5,
				expectedBytes: []byte("hoge\nfuga\n"),
				expectedError: nil,
				desc:          "KinesisStreamsRecordSizeLimit test: exclude too long line",
			},
			&ReadBytesByLineTestCase{
				content:       "",
				start:         5 + 1024*1024 + 6,
				maxLineSize:   1024 * 1024,
				expectedN:     0,
				expectedBytes: []byte{},
				expectedError: nil,
				desc:          "KinesisStreamsRecordSizeLimit test: read no lines",
			},
		},
	}

	dir, err := ioutil.TempDir("", "file_reader")
	checkErr(err)
	defer os.RemoveAll(dir)

	var reader *FileReader
	for i, cases := range testCases {
		fn := filepath.Join(dir, fmt.Sprintf("test%d.log", i))
		f, err := os.OpenFile(
			fn,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC,
			FileOpenPermission,
		)
		checkErr(err)

		pinode := getInode(fn)
		reader = newFileReader(fn, *pinode)

		for _, c := range cases {
			f.WriteString(c.content)
			reader.MaxLineSize = c.maxLineSize
			n, bytes, err := reader.readBytesByLine(c.start)
			assert.Equal(t, c.expectedN, n, fmt.Sprintf("%s: test size", c.desc))
			assert.Equal(t, c.expectedBytes, bytes, fmt.Sprintf("%s: test lines", c.desc))
			assert.Equal(t, c.expectedError, err, fmt.Sprintf("%s: test error", c.desc))
		}
	}
	reader.Close()
}

func TestReadBytesByLineWithSizeOver(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_reader")
	checkErr(err)
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test.log")
	f, err := os.OpenFile(
		fn,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC,
		FileOpenPermission,
	)
	checkErr(err)

	str := strings.Repeat("a", 1024*1024) + "\n"
	f.WriteString(str)

	pinode := getInode(fn)
	reader := newFileReader(fn, *pinode)

	backupPath := filepath.Join(dir, "unputtable")
	backupIO, err := os.OpenFile(
		backupPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC,
		0644,
	)
	checkErr(err)
	reader.backupIO = backupIO

	n, b, err := reader.readBytesByLine(0)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, []byte{}, b)
	assert.NoError(t, err)

	content, err := ioutil.ReadFile(backupPath)
	checkErr(err)
	assert.Equal(t, []byte(str), content)
}

func TestReadLinesWithError(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_reader")
	checkErr(err)
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test.log")
	f, err := os.OpenFile(
		fn,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC,
		FileOpenPermission,
	)
	checkErr(err)

	str := "hoge\n"
	f.WriteString(str)

	pinode := getInode(fn)
	reader := newFileReader(fn, *pinode)

	chunk, err := reader.ReadLines()
	assert.NotNil(t, chunk)
	assert.NoError(t, err)

	os.Remove(fn)

	chunk, err = reader.ReadLines()
	assert.Nil(t, chunk)
	assert.Error(t, err)
}

func newFileReader(path string, inode uint64) *FileReader {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, FileOpenPermission)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	io := file.NewFileWrapper(f)

	return &FileReader{
		path:        path,
		inode:       inode,
		pos:         0,
		io:          io,
		backupIO:    nil,
		lifetimer:   &lifetimer.LifeTimer{},
		MaxLineSize: DefaultMaxLineSize,
	}
}

func getInode(path string) *uint64 {
	stat, err := os.Stat(path)
	if err != nil {
		return nil
	}
	sysstat, _ := stat.Sys().(*syscall.Stat_t)
	return &sysstat.Ino
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
