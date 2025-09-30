package spicyreload

import (
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func addFoldersToWatch(watcher *fsnotify.Watcher, pathList []string) {
	for _, path := range pathList {
		walkAndRegister(path, watcher)
	}
}

func walkAndRegister(path string, watcher *fsnotify.Watcher) {
	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("error in walkfunc", "err", err.Error())
			return nil
		}

		if d.IsDir() {
			err := watcher.Add(path)
			if err != nil {
				slog.Error("error adding path to watcher", "err", err.Error())
			}
		}
		return nil
	})
}
