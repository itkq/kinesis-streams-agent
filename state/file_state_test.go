package state

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAndDumpFileState(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_state")
	assert.Equal(t, nil, err)
	defer func() {
		err := os.RemoveAll(dir)
		assert.Equal(t, nil, err)
	}()
	fn := filepath.Join(dir, "test.state")

	// empty test
	fileState, err := LoadFromJSON(fn)
	assert.NotEqual(t, nil, fileState)
	assert.Equal(t, nil, err)

	content := `
{
	"6406163": {
		"path": "/tmp/kinesis-agent-go/test1.log",
		"pos": 5119,
		"send_ranges": [
			{
				"begin": 0,
				"end": 5119
			}
		]
	},
	"6421632": {
		"path": "/tmp/kinesis-agent-go/test1.log",
		"pos": 3071,
		"send_ranges": [
			{
				"begin": 0,
				"end": 1023
			},
			{
				"begin": 2048,
				"end": 3071
			}
		]
	},
	"6414379": {
		"path": "/tmp/kinesis-agent-go/test2.log",
		"pos": 3071,
		"send_ranges": [
			{
				"begin": 1024,
				"end": 2047
			}
		]
	}
}`

	ioutil.WriteFile(fn, []byte(content), 0644)

	fileState, err = LoadFromJSON(fn)
	assert.NotEqual(t, nil, fileState)
	assert.Equal(t, nil, err)

	err = fileState.DumpToJSON()
	assert.Equal(t, nil, err)

	fileState2, err := LoadFromJSON(fn)
	assert.NotEqual(t, nil, fileState)
	assert.Equal(t, nil, err)
	assert.Equal(t, fileState, fileState2)
}

func TestGetAndCreateReaderState(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_state")
	assert.Equal(t, nil, err)
	defer func() {
		err := os.RemoveAll(dir)
		assert.Equal(t, nil, err)
	}()
	fn := filepath.Join(dir, "test.state")
	fs := NewFileState(fn)

	dummyPath := "hoge.log"
	var dummyInode uint64 = 1000000

	rs := fs.GetReaderState(dummyInode)
	assert.Equal(t, (*ReaderState)(nil), rs)
	rs = fs.CreateReaderState(dummyInode, dummyPath)
	assert.NotEqual(t, (*ReaderState)(nil), rs)
	assert.Equal(t, dummyPath, rs.Path)
}

func TestFileStateUpdate(t *testing.T) {
	state := NewFileState("dummy")

	var dummyInode uint64 = 10000
	si1 := &SendInfo{
		Inode: dummyInode,
		ReadRange: &FileReadRange{
			Begin: 0,
			End:   10,
		},
		Succeeded: true,
	}
	state.Update(si1)
	rstate := state.readerStates[dummyInode]
	assert.Equal(t, si1.ReadRange.End, rstate.Pos)

	si2 := &SendInfo{
		Inode: dummyInode,
		ReadRange: &FileReadRange{
			Begin: 10,
			End:   20,
		},
		Succeeded: false,
	}
	state.Update(si2)
	rstate.Compact()
	assert.Equal(t, si2.ReadRange.End, rstate.Pos)
	assert.Equal(t, si1.ReadRange, rstate.SendRanges[0])

	si3 := &SendInfo{
		Inode: dummyInode,
		ReadRange: &FileReadRange{
			Begin: 20,
			End:   30,
		},
		Succeeded: true,
	}
	state.Update(si3)
	rstate.Compact()
	assert.Equal(t, si3.ReadRange.End, rstate.Pos)
	assert.Equal(t, si1.ReadRange, rstate.SendRanges[0])
	assert.Equal(t, si3.ReadRange, rstate.SendRanges[1])

	si2.Succeeded = true
	state.Update(si2)
	rstate.Compact()
	assert.Equal(t, si3.ReadRange.End, rstate.Pos)
	assert.Equal(
		t,
		&FileReadRange{
			Begin: si1.ReadRange.Begin,
			End:   si3.ReadRange.End,
		},
		rstate.SendRanges[0],
	)
}

func TestFileStateCompact(t *testing.T) {

}

func TestReaderStateCompact(t *testing.T) {
	rstate := NewReaderState()
	r1 := &FileReadRange{
		Begin: 0,
		End:   10,
	}
	rstate.AddSendRange(r1)
	rstate.Compact()
	assert.Equal(t, r1, rstate.SendRanges[0])

	r2 := &FileReadRange{
		Begin: 20,
		End:   30,
	}
	rstate.AddSendRange(r2)
	rstate.Compact()
	assert.Equal(t, r1, rstate.SendRanges[0])
	assert.Equal(t, r2, rstate.SendRanges[1])

	r3 := &FileReadRange{
		Begin: 10,
		End:   20,
	}
	rstate.AddSendRange(r3)
	rstate.Compact()
	assert.Equal(
		t,
		&FileReadRange{
			Begin: 0,
			End:   30,
		},
		rstate.SendRanges[0],
	)
}

func TestLeakedRanges(t *testing.T) {
	type testCase struct {
		rstate         *ReaderState
		expectedRanges []*FileReadRange
	}

	testCases := []*testCase{
		&testCase{
			rstate: &ReaderState{
				Pos:        0,
				SendRanges: []*FileReadRange{},
			},
			expectedRanges: []*FileReadRange{},
		},
		&testCase{
			rstate: &ReaderState{
				Pos:        4,
				SendRanges: []*FileReadRange{},
			},
			expectedRanges: []*FileReadRange{
				&FileReadRange{
					Begin: 0,
					End:   4,
				},
			},
		},
		&testCase{
			rstate: &ReaderState{
				Pos: 4,
				SendRanges: []*FileReadRange{
					&FileReadRange{
						Begin: 0,
						End:   4,
					},
				},
			},
			expectedRanges: []*FileReadRange{},
		},
		&testCase{
			rstate: &ReaderState{
				Pos: 4,
				SendRanges: []*FileReadRange{
					&FileReadRange{
						Begin: 2,
						End:   4,
					},
				},
			},
			expectedRanges: []*FileReadRange{
				&FileReadRange{
					Begin: 0,
					End:   2,
				},
			},
		},
		&testCase{
			rstate: &ReaderState{
				Pos: 6,
				SendRanges: []*FileReadRange{
					&FileReadRange{
						Begin: 0,
						End:   2,
					},
					&FileReadRange{
						Begin: 4,
						End:   6,
					},
				},
			},
			expectedRanges: []*FileReadRange{
				&FileReadRange{
					Begin: 2,
					End:   4,
				},
			},
		},
		&testCase{
			rstate: &ReaderState{
				Pos: 6,
				SendRanges: []*FileReadRange{
					&FileReadRange{
						Begin: 0,
						End:   4,
					},
				},
			},
			expectedRanges: []*FileReadRange{
				&FileReadRange{
					Begin: 4,
					End:   6,
				},
			},
		},
	}

	for _, c := range testCases {
		assert.Equal(t, c.expectedRanges, c.rstate.LeakedRanges())
	}
}
