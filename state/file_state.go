package state

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
	"syscall"
)

const (
	FileOpenPermission = 0644
)

type SendInfo struct {
	Inode     uint64
	ReadRange *FileReadRange
	Succeeded bool
}

type FileState struct {
	*sync.Mutex
	// state file path
	path string
	// inode -> *FileReaderState
	readerStates map[uint64]*ReaderState
}

func NewFileState(path string) *FileState {
	return &FileState{
		new(sync.Mutex),
		path,
		make(map[uint64]*ReaderState),
	}
}

func LoadFromJSON(path string) (*FileState, error) {
	stat, err := os.Stat(path)
	if stat == nil {
		log.Println("info: create state file")
		f, err := os.OpenFile(
			path,
			os.O_CREATE|os.O_RDWR|os.O_SYNC,
			FileOpenPermission,
		)
		if err != nil {
			return nil, err
		}
		f.Close()

		s := NewFileState(path)
		s.DumpToJSON()

		return s, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var readerStates map[uint64]*ReaderState
	err = json.Unmarshal(bytes, &readerStates)
	if err != nil {
		return nil, err
	}

	return &FileState{
		new(sync.Mutex),
		path,
		readerStates,
	}, nil
}

func (s *FileState) DumpToJSON() error {
	s.Lock()
	defer s.Unlock()

	s.Compact()

	b, err := json.Marshal(s.readerStates)
	if err != nil {
		return err
	}

	io, err := os.OpenFile(
		s.path,
		os.O_WRONLY|os.O_CREATE|os.O_SYNC,
		FileOpenPermission,
	)
	if err != nil {
		return err
	}
	out := new(bytes.Buffer)
	json.Indent(out, b, "", "    ")

	io.Truncate(0)
	_, err = io.WriteString(out.String())

	return err
}

func (s *FileState) Compact() {
	for inode, rs := range s.readerStates {
		if len(rs.SendRanges) == 0 {
			continue
		}
		ainode := rs.GetActualInode()
		lastRange := rs.SendRanges[len(rs.SendRanges)-1]
		if rs.Pos == lastRange.End && (ainode == nil || inode != *ainode) {
			delete(s.readerStates, inode)
		}
	}
}

func (s *FileState) Update(si *SendInfo) {
	s.Lock()
	defer s.Unlock()

	rs, ok := s.readerStates[si.Inode]
	if !ok {
		rs = NewReaderState()
		s.readerStates[si.Inode] = rs
	}

	if si.Succeeded {
		rs.AddSendRange(si.ReadRange)
		rs.Compact()
	}
	rs.UpdatePos(si.ReadRange)
}

func (s *FileState) GetReaderState(inode uint64) *ReaderState {
	s.Lock()
	defer s.Unlock()

	return s.readerStates[inode]
}

func (s *FileState) CreateReaderState(inode uint64, path string) *ReaderState {
	rstate := NewReaderState()
	rstate.Path = path
	s.readerStates[inode] = rstate

	return s.readerStates[inode]
}

type ReaderState struct {
	Pos        int64            `json:"pos,requied"`
	Path       string           `json:"path,required"`
	SendRanges []*FileReadRange `json:"send_ranges,required"`
}

func NewReaderState() *ReaderState {
	return &ReaderState{
		Pos:        0,
		SendRanges: make([]*FileReadRange, 0),
	}
}

func (s *ReaderState) AddSendRange(r *FileReadRange) {
	s.SendRanges = append(s.SendRanges, r)
}

func (s *ReaderState) UpdatePos(r *FileReadRange) {
	if s.Pos < r.End {
		s.Pos = r.End
	}
}

func (s *ReaderState) Compact() {
	s.SortSendRanges()

	if len(s.SendRanges) <= 1 {
		return
	}

	newsr := []*FileReadRange{
		s.SendRanges[0],
	}
	for i := 1; i < len(s.SendRanges); i++ {
		last := newsr[len(newsr)-1]
		r := s.SendRanges[i]
		if last.End == r.Begin {
			last.End = r.End
		} else {
			newsr = append(newsr, r)
		}
	}

	s.SendRanges = newsr
}

func (s *ReaderState) SortSendRanges() {
	sort.Slice(s.SendRanges, func(i, j int) bool {
		return s.SendRanges[i].Begin < s.SendRanges[j].Begin
	})
}

func (s *ReaderState) LeakedRanges() []*FileReadRange {
	leakedRanges := make([]*FileReadRange, 0)

	if len(s.SendRanges) == 0 {
		if s.Pos != 0 {
			leakedRanges = append(leakedRanges, &FileReadRange{
				Begin: 0,
				End:   s.Pos,
			})
		}
		return leakedRanges
	}

	// head
	if s.SendRanges[0].Begin != 0 {
		leakedRanges = append(leakedRanges, &FileReadRange{
			Begin: 0,
			End:   s.SendRanges[0].Begin,
		})
	}

	// intermediate
	for i := 0; i < len(s.SendRanges)-1; i++ {
		if s.SendRanges[i].End != s.SendRanges[i+1].Begin {
			leakedRanges = append(leakedRanges, &FileReadRange{
				Begin: s.SendRanges[i].End,
				End:   s.SendRanges[i+1].Begin,
			})
		}
	}

	// tail
	if s.SendRanges[len(s.SendRanges)-1].End != s.Pos {
		leakedRanges = append(leakedRanges, &FileReadRange{
			Begin: s.SendRanges[len(s.SendRanges)-1].End,
			End:   s.Pos,
		})
	}

	return leakedRanges
}

func (s *ReaderState) GetActualInode() *uint64 {
	info, err := os.Stat(s.Path)
	if err != nil {
		return nil
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		return &stat.Ino
	}

	return nil
}

type FileReadRange struct {
	Begin int64 `json:"begin,required"`
	End   int64 `json:"end,required"`
}

func (r *FileReadRange) Len() int64 {
	return r.End - r.Begin
}
