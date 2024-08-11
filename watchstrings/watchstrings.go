package watchstrings

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	watchModeCache  = map[string]watchState{}
	watchStateMutex sync.Mutex
)

type watchState struct {
	modTime time.Time
	strings *[]string
}

func Watch(ptr *[]string) {
	_, path, _, _ := runtime.Caller(1)
	if !strings.HasSuffix(path, "_templ.go") {
		panic("templ strings: Watch can only be called from _templ.go files")
	}
	txtFilePath := strings.Replace(path, "_templ.go", "_templ.txt", 1)

	// Set cache.
	watchStateMutex.Lock()
	defer watchStateMutex.Unlock()
	state := watchModeCache[txtFilePath]
	state.strings = ptr
	watchModeCache[txtFilePath] = state

	//TODO: Add a file watcher to update the cache, but for now, poll.
	go func() {
		for {
			time.Sleep(time.Second)
			_, err := getStringsFile(txtFilePath)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
}

func Get(index int) (string, error) {
	_, path, _, _ := runtime.Caller(1)
	if !strings.HasSuffix(path, "_templ.go") {
		return "", errors.New("templ strings: Get can only be called from _templ.go files")
	}
	txtFilePath := strings.Replace(path, "_templ.go", "_templ.txt", 1)

	literals, err := getStringsFile(txtFilePath)
	if err != nil {
		return "", fmt.Errorf("templ strings: failed to cache strings: %w", err)
	}

	if index > len(literals) {
		return "", fmt.Errorf("templ strings: failed to find line %d in %s", index, txtFilePath)
	}

	unquoted, err := strconv.Unquote(`"` + literals[index] + `"`)
	if err != nil {
		return "", fmt.Errorf("templ strings: failed to unquote %s: %w", literals[index-1], err)
	}

	return unquoted, nil
}

func getStringsFile(txtFilePath string) ([]string, error) {
	watchStateMutex.Lock()
	defer watchStateMutex.Unlock()

	state, isCached := watchModeCache[txtFilePath]
	if isCached && (time.Since(state.modTime) < time.Millisecond*100) {
		return *state.strings, nil
	}

	// See if the file has changed.
	info, err := os.Stat(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}
	if isCached && !info.ModTime().After(state.modTime) {
		return *state.strings, nil
	}

	// Get file contents from disk.
	fc, err := os.ReadFile(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to get file contents: %w", err)
	}
	literals := strings.Split(string(fc), "\n")

	// Update cache.
	state.modTime = info.ModTime()
	if state.strings == nil {
		state.strings = &literals
	}
	*state.strings = literals
	watchModeCache[txtFilePath] = state

	return literals, nil
}
