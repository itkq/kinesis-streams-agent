package fswatcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Fswatcher watches filesystem event using fsnotify.Watcher
type Fswatcher struct {
	*fsnotify.Watcher
	WatchingPaths []string
	watchingDir   map[string]bool
}

func NewFswatcher() (*Fswatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Fswatcher{
		watcher,
		make([]string, 0),
		make(map[string]bool),
	}, nil
}

func (w *Fswatcher) RegisterPaths(paths []string) {
	for _, path := range paths {
		var dir string

		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			dir = path
		} else {
			dir = filepath.Dir(path)
		}

		if _, ok := w.watchingDir[dir]; !ok {
			w.Add(dir)
			w.watchingDir[dir] = true
			log.Println("info: watch dir", dir)
		}
	}

	w.WatchingPaths = append(w.WatchingPaths, paths...)
}

func (w *Fswatcher) IsCreatedEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create
}

func (w *Fswatcher) ShouldWatchEvent(ev fsnotify.Event) bool {
	path := ev.Name

	for _, p := range w.ExpandPaths() {
		if p == path {
			return true
		}
	}

	return false
}

func (w *Fswatcher) ExpandPaths() []string {
	paths := make([]string, 0)
	for _, path := range w.WatchingPaths {
		var globpath string

		if strings.Index(path, "*") != -1 {
			globpath = path
		} else {
			info, err := os.Stat(path)
			if err != nil {
				// file does not exist currently
				continue
			}
			if info.IsDir() {
				globpath = filepath.Join(path, "*")
			} else {
				paths = append(paths, path)
				continue
			}
		}

		matches, err := filepath.Glob(globpath)
		if err != nil {
			log.Println("error:", err)
			continue
		}

		for _, p := range matches {
			paths = append(paths, p)
		}
	}

	return paths
}
