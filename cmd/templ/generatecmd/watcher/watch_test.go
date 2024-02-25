package watcher

import (
	"context"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchDebouncesDuplicates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rw := &RecursiveWatcher{
		ctx: ctx,
		w: &fsnotify.Watcher{
			Events: make(chan fsnotify.Event),
		},
		Events: make(chan fsnotify.Event, 2),
		timers: make(map[string]*time.Timer),
	}
	go func() {
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		cancel()
		close(rw.w.Events)
	}()
	rw.loop()
	count := 0
	exp := time.After(300 * time.Millisecond)
	for {
		select {
		case <-rw.Events:
			count++
		case <-exp:
			if count != 1 {
				t.Errorf("expected 1 event, got %d", count)
			}
			return
		}
	}
}

func TestWatchDoesNotDebounceDifferentEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rw := &RecursiveWatcher{
		ctx: ctx,
		w: &fsnotify.Watcher{
			Events: make(chan fsnotify.Event),
		},
		Events: make(chan fsnotify.Event, 2),
		timers: make(map[string]*time.Timer),
	}
	go func() {
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		rw.w.Events <- fsnotify.Event{Name: "test2.templ"}
		cancel()
		close(rw.w.Events)
	}()
	rw.loop()
	count := 0
	exp := time.After(300 * time.Millisecond)
	for {
		select {
		case <-rw.Events:
			count++
		case <-exp:
			if count != 2 {
				t.Errorf("expected 2 event, got %d", count)
			}
			return
		}
	}
}

func TestWatchDoesNotDebounceSeparateEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rw := &RecursiveWatcher{
		ctx: ctx,
		w: &fsnotify.Watcher{
			Events: make(chan fsnotify.Event),
		},
		Events: make(chan fsnotify.Event, 2),
		timers: make(map[string]*time.Timer),
	}
	go func() {
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		<-time.After(200 * time.Millisecond)
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		cancel()
		close(rw.w.Events)
	}()
	rw.loop()
	count := 0
	exp := time.After(500 * time.Millisecond)
	for {
		select {
		case <-rw.Events:
			count++
		case <-exp:
			if count != 2 {
				t.Errorf("expected 2 event, got %d", count)
			}
			return
		}
	}
}
