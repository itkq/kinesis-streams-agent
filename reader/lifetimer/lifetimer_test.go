package lifetimer

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLifeTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", "lifetimer")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "test.txt")
	_, err = os.OpenFile(path, os.O_CREATE, 0644) // create file
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	stat, _ := os.Stat(path)
	if stat == nil {
		log.Println(err)
		os.Exit(1)
	}
	sysstat, _ := stat.Sys().(*syscall.Stat_t)
	inode := sysstat.Ino

	timer := NewLifeTimer(path, inode)
	timer.LifeTime = 100 * time.Millisecond

	assert.Equal(t, false, timer.ShouldDie())
	assert.Equal(t, false, timer.ShouldDie())

	renamed_path := filepath.Join(dir, "renamed.txt")
	os.Rename(path, renamed_path)
	assert.Equal(t, false, timer.ShouldDie())

	time.Sleep(timer.LifeTime * 2)
	assert.Equal(t, true, timer.ShouldDie())
	assert.Equal(t, true, timer.ShouldDie())
}
