package lifetimer

import (
	"os"
	"syscall"
	"time"
)

const (
	DefaultLifeTime = 5 * time.Second
)

type LifeTimer struct {
	LifeTime    time.Duration
	started     bool
	path        string
	inode       uint64
	lastUpdated time.Time
}

func NewLifeTimer(path string, inode uint64) *LifeTimer {
	return &LifeTimer{
		LifeTime: DefaultLifeTime,
		started:  false,
		path:     path,
		inode:    inode,
	}
}

func (t *LifeTimer) ShouldDie() bool {
	if !t.started {
		t.lastUpdated = time.Now()
		t.started = true
		return false
	}

	return t.moved() && time.Now().Sub(t.lastUpdated) > t.LifeTime
}

func (t *LifeTimer) moved() bool {
	stat, _ := os.Stat(t.path)
	if stat == nil {
		return true
	}
	sysstat, _ := stat.Sys().(*syscall.Stat_t)

	return t.inode != sysstat.Ino
}
