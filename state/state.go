package state

type State interface {
	DumpToJSON() error
	GetReaderState(inode uint64) *ReaderState
	CreateReaderState(inode uint64, path string) *ReaderState
	Update(info *SendInfo)
}

type DummyState struct{}

func (s *DummyState) DumpToJSON() error {
	return nil
}

func (s *DummyState) GetReaderState(inode uint64) *ReaderState {
	return NewReaderState()
}

func (s *DummyState) CreateReaderState(inode uint64, path string) *ReaderState {
	return NewReaderState()
}

func (s *DummyState) Update(info *SendInfo) {}
