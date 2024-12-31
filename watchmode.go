package templ

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WriteWatchModeString is used when rendering templates in development mode.
// the generator would have written non-go code to the _templ.txt file, which
// is then read by this function and written to the output.
//
// Deprecated: since templ v0.3.x generated code uses WriteString.
func WriteWatchModeString(w io.Writer, lineNum int) error {
	_, path, _, _ := runtime.Caller(1)
	if !strings.HasSuffix(path, "_templ.go") {
		return errors.New("templ: WriteWatchModeString can only be called from _templ.go")
	}
	txtFilePath := strings.Replace(path, "_templ.go", "_templ.txt", 1)

	literals, err := getWatchedStrings(txtFilePath)
	if err != nil {
		return fmt.Errorf("templ: failed to cache strings: %w", err)
	}

	if lineNum > len(literals) {
		return fmt.Errorf("templ: failed to find line %d in %s", lineNum, txtFilePath)
	}

	s, err := strconv.Unquote(`"` + literals[lineNum-1] + `"`)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

var (
	watchModeCache  = map[string]watchState{}
	watchStateMutex sync.Mutex
)

type watchState struct {
	modTime time.Time
	strings []string
}

func getWatchedStrings(txtFilePath string) ([]string, error) {
	watchStateMutex.Lock()
	defer watchStateMutex.Unlock()

	state, cached := watchModeCache[txtFilePath]
	if !cached {
		return cacheStrings(txtFilePath)
	}

	if time.Since(state.modTime) < time.Millisecond*100 {
		return state.strings, nil
	}

	info, err := os.Stat(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}

	if !info.ModTime().After(state.modTime) {
		return state.strings, nil
	}

	return cacheStrings(txtFilePath)
}

func cacheStrings(txtFilePath string) ([]string, error) {
	txtFile, err := os.Open(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to open %s: %w", txtFilePath, err)
	}
	defer txtFile.Close()

	info, err := txtFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}

	all, err := io.ReadAll(txtFile)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to read %s: %w", txtFilePath, err)
	}

	literals := strings.Split(string(all), "\n")
	watchModeCache[txtFilePath] = watchState{
		modTime: info.ModTime(),
		strings: literals,
	}

	return literals, nil
}
