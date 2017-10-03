package filewatcher

// func (w *FileWatcher) Export() api.Metrics {
// 	watchingDirs := make([]string, 0, len(w.fswatcher.watchingDir))
// 	for dir := range w.fswatcher.watchingDir {
// 		watchingDirs = append(watchingDirs, dir)
// 	}
// 	watchingFiles := make([]string, 0, len(w.fswatcher.watchingFile))
// 	for file := range w.fswatcher.watchingFile {
// 		watchingFiles = append(watchingFiles, file)
// 	}

// 	return &metrics{
// 		WatchingDirs:  watchingDirs,
// 		WatchingFiles: watchingFiles,
// 	}
// }

// func (w *FileWatcher) Endpoint() string {
// 	return "/file_watcher"
// }

// type metrics struct {
// 	WatchingDirs  []string `json:"watching_directories"`
// 	WatchingFiles []string `json:"watching_files"`
// }

// func (m *metrics) ToJSON() ([]byte, error) {
// 	return json.Marshal(m)
// }
