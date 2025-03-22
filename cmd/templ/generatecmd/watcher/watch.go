package watcher

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/a-h/templ/internal/skipdir"
	"github.com/fsnotify/fsnotify"
)

func Recursive(
	ctx context.Context,
	path string,
	watchPattern *regexp.Regexp,
	out chan fsnotify.Event,
	errors chan error,
) (w *RecursiveWatcher, err error) {
	fsnw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w = NewRecursiveWatcher(ctx, fsnw, watchPattern, out, errors)
	go w.loop()
	return w, w.Add(path)
}

func NewRecursiveWatcher(ctx context.Context, w *fsnotify.Watcher, watchPattern *regexp.Regexp, events chan fsnotify.Event, errors chan error) *RecursiveWatcher {
	return &RecursiveWatcher{
		ctx:          ctx,
		w:            w,
		WatchPattern: watchPattern,
		Events:       events,
		Errors:       errors,
		timers:       make(map[timerKey]*time.Timer),
	}
}

// WalkFiles walks the file tree rooted at path, sending a Create event for each
// file it encounters.
func WalkFiles(ctx context.Context, path string, watchPattern *regexp.Regexp, out chan fsnotify.Event) (err error) {
	rootPath := path
	fileSystem := os.DirFS(rootPath)
	return fs.WalkDir(fileSystem, ".", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		absPath, err := filepath.Abs(filepath.Join(rootPath, path))
		if err != nil {
			return nil
		}
		if info.IsDir() && skipdir.ShouldSkip(absPath) {
			return filepath.SkipDir
		}
		if !watchPattern.MatchString(absPath) {
			return nil
		}
		out <- fsnotify.Event{
			Name: absPath,
			Op:   fsnotify.Create,
		}
		return nil
	})
}

type RecursiveWatcher struct {
	ctx          context.Context
	w            *fsnotify.Watcher
	WatchPattern *regexp.Regexp
	Events       chan fsnotify.Event
	Errors       chan error
	timerMu      sync.Mutex
	timers       map[timerKey]*time.Timer
}

type timerKey struct {
	name string
	op   fsnotify.Op
}

func timerKeyFromEvent(event fsnotify.Event) timerKey {
	return timerKey{
		name: event.Name,
		op:   event.Op,
	}
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
			if !w.WatchPattern.MatchString(event.Name) {
				continue
			}
			tk := timerKeyFromEvent(event)
			w.timerMu.Lock()
			t, ok := w.timers[tk]
			w.timerMu.Unlock()
			if !ok {
				t = time.AfterFunc(100*time.Millisecond, func() {
					w.Events <- event
				})
				w.timerMu.Lock()
				w.timers[tk] = t
				w.timerMu.Unlock()
				continue
			}
			t.Reset(100 * time.Millisecond)
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
		if skipdir.ShouldSkip(dir) {
			return filepath.SkipDir
		}
		return w.w.Add(dir)
	})
}
