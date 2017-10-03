package fswatcher

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFswatcher(t *testing.T) {
	dir, err := ioutil.TempDir("", "fswatcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	fswatcher, err := NewFswatcher()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	paths := []string{
		filepath.Join(dir, "test*.log"),
		filepath.Join(dir, "test.log"),
	}
	fswatcher.RegisterPaths(paths)
	assert.Equal(t, paths, fswatcher.WatchingPaths)
	assert.Equal(t, true, fswatcher.watchingDir[dir])

	fn1 := filepath.Join(dir, "test1.log")
	go func() {
		_, err = os.OpenFile(fn1, os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()

	ev := <-fswatcher.Events
	assert.Equal(t, fn1, ev.Name)

	fn2 := filepath.Join(dir, "test.log")
	go func() {
		_, err = os.OpenFile(fn2, os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()

	ev = <-fswatcher.Events
	assert.Equal(t, fn2, ev.Name)
}

func TestExpandPaths(t *testing.T) {
	dir1, err := ioutil.TempDir("", "fswatcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir1)
	dir2, err := ioutil.TempDir("", "fswatcher")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir2)

	fswatcher, err := NewFswatcher()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fswatcher.RegisterPaths([]string{
		filepath.Join(dir1, "test*.log"),
		dir2,
	})

	paths := map[string]struct{}{
		filepath.Join(dir1, "test1.log"): {},
		filepath.Join(dir2, "hoge"):      {},
		filepath.Join(dir2, "fuga"):      {},
	}
	for p := range paths {
		createFile(p)
	}
	ignorePaths := map[string]struct{}{
		filepath.Join(dir1, "hoge.log"): {},
		dir1: {},
		dir2: {},
	}
	for p := range ignorePaths {
		createFile(p)
	}

	var ok bool
	for _, p := range fswatcher.ExpandPaths() {
		_, ok = paths[p]
		assert.Equal(t, true, ok)
		_, ok = ignorePaths[p]
		assert.Equal(t, false, ok)
	}
}

func createFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}
