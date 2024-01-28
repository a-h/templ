package watcher

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func Recursive(ctx context.Context, path string, out chan fsnotify.Event, errors chan error) (w *RecursiveWatcher, err error) {
	fsnw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w = &RecursiveWatcher{
		ctx:    ctx,
		w:      fsnw,
		Events: out,
		Errors: errors,
	}
	go w.loop()
	return w, w.Add(path)
}

// WalkFiles walks the file tree rooted at path, sending a Create event for each
// file it encounters.
func WalkFiles(ctx context.Context, path string, out chan fsnotify.Event) (err error) {
	return filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && shouldSkipDir(path) {
			return filepath.SkipDir
		}
		if !shouldIncludeFile(path) {
			return nil
		}
		out <- fsnotify.Event{
			Name: path,
			Op:   fsnotify.Create,
		}
		return nil
	})
}

func shouldIncludeFile(name string) bool {
	if strings.HasSuffix(name, ".templ") {
		return true
	}
	if strings.HasSuffix(name, "_templ.go") {
		return true
	}
	if strings.HasSuffix(name, "_templ.txt") {
		return true
	}
	return false
}

type RecursiveWatcher struct {
	ctx    context.Context
	w      *fsnotify.Watcher
	Events chan fsnotify.Event
	Errors chan error
}

func (w *RecursiveWatcher) Close() error {
	return w.w.Close()
}

func (w *RecursiveWatcher) loop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if err := w.Add(event.Name); err != nil {
					w.Errors <- err
				}
			}
			// Only notify on templ related files.
			if !shouldIncludeFile(event.Name) {
				continue
			}
			w.Events <- event
		case err, ok := <-w.w.Errors:
			if !ok {
				return
			}
			w.Errors <- err
		}
	}
}

func (w *RecursiveWatcher) Add(dir string) error {
	return filepath.WalkDir(dir, func(dir string, info os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if shouldSkipDir(dir) {
			return filepath.SkipDir
		}
		return w.w.Add(dir)
	})
}

func shouldSkipDir(dir string) bool {
	if dir == "." {
		return false
	}
	if dir == "vendor" || dir == "node_modules" {
		return true
	}
	_, name := path.Split(dir)
	// These directories are ignored by the Go tool.
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return true
	}
	return false
}
