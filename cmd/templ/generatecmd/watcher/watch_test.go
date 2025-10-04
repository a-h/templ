package watcher

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchDebouncesDuplicates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan fsnotify.Event, 2)
	errors := make(chan error)
	watchPattern, err := regexp.Compile(".*")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to compile watch pattern: %w", err))
	}
	rw, err := Recursive(ctx, watchPattern, nil, events, errors)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create recursive watcher: %w", err))
	}
	go func() {
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
	}()
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
			cancel()
			if err := rw.Close(); err != nil {
				t.Errorf("unexpected error closing watcher: %v", err)
			}
			return
		}
	}
}

func TestWatchDoesNotDebounceDifferentEvents(t *testing.T) {
	tests := []struct {
		event1 fsnotify.Event
		event2 fsnotify.Event
	}{
		// Different files
		{fsnotify.Event{Name: "test.templ"}, fsnotify.Event{Name: "test2.templ"}},
		// Different operations
		{
			fsnotify.Event{Name: "test.templ", Op: fsnotify.Create},
			fsnotify.Event{Name: "test.templ", Op: fsnotify.Write},
		},
		// Different operations and files
		{
			fsnotify.Event{Name: "test.templ", Op: fsnotify.Create},
			fsnotify.Event{Name: "test2.templ", Op: fsnotify.Write},
		},
	}
	for _, test := range tests {
		ctx, cancel := context.WithCancel(context.Background())
		events := make(chan fsnotify.Event, 2)
		errors := make(chan error)
		watchPattern, err := regexp.Compile(".*")
		if err != nil {
			t.Fatal(fmt.Errorf("failed to compile watch pattern: %w", err))
		}
		rw, err := Recursive(ctx, watchPattern, nil, events, errors)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create recursive watcher: %w", err))
		}
		go func() {
			rw.w.Events <- test.event1
			rw.w.Events <- test.event2
		}()
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
				cancel()
				if err := rw.Close(); err != nil {
					t.Errorf("unexpected error closing watcher: %v", err)
				}
				return
			}
		}
	}
}

func TestWatchDoesNotDebounceSeparateEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan fsnotify.Event, 2)
	errors := make(chan error)
	watchPattern, err := regexp.Compile(".*")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to compile watch pattern: %w", err))
	}
	rw, err := Recursive(ctx, watchPattern, nil, events, errors)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create recursive watcher: %w", err))
	}
	go func() {
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
		<-time.After(200 * time.Millisecond)
		rw.w.Events <- fsnotify.Event{Name: "test.templ"}
	}()
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
			cancel()
			if err := rw.Close(); err != nil {
				t.Errorf("unexpected error closing watcher: %v", err)
			}
			return
		}
	}
}

func TestWatchIgnoresFilesMatchingIgnorePattern(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan fsnotify.Event, 2)
	errors := make(chan error)
	watchPattern, err := regexp.Compile(".*")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to compile watch pattern: %w", err))
	}
	ignorePattern, err := regexp.Compile(`ignore\.templ$`)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to compile ignore pattern: %w", err))
	}

	rw, err := Recursive(ctx, watchPattern, ignorePattern, events, errors)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create recursive watcher: %w", err))
	}

	go func() {
		// This should be ignored
		rw.w.Events <- fsnotify.Event{Name: "file.ignore.templ"}
		// This should pass
		rw.w.Events <- fsnotify.Event{Name: "file.keep.templ"}
	}()

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
			cancel()
			if err := rw.Close(); err != nil {
				t.Errorf("unexpected error closing watcher: %v", err)
			}
			return
		}
	}
}

func TestIgnorePatternTakesPrecedenceOverWatchPattern(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan fsnotify.Event, 2)
	errors := make(chan error)
	watchPattern := regexp.MustCompile(`.*\.templ$`)
	ignorePattern := regexp.MustCompile(`.*\.ignore\.templ$`)

	rw, err := Recursive(ctx, watchPattern, ignorePattern, events, errors)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		rw.w.Events <- fsnotify.Event{Name: "file.ignore.templ"}
	}()

	exp := time.After(300 * time.Millisecond)
	select {
	case <-rw.Events:
		t.Errorf("expected no events because ignore should win")
	case <-exp:
		cancel()
		_ = rw.Close()
	}
}
