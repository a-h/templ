package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var developmentMode = os.Getenv("TEMPL_DEV_MODE") == "true"

var stringLoaderOnce = sync.OnceValue(func() *StringLoader {
	return NewStringLoader(os.Getenv("TEMPL_DEV_MODE_WATCH_ROOT"))
})

// WriteString writes the string to the writer. If development mode is enabled
// s is replaced with the string at the index in the _templ.txt file.
func WriteString(w io.Writer, index int, s string) (err error) {
	if developmentMode {
		_, path, _, _ := runtime.Caller(1)
		if !strings.HasSuffix(path, "_templ.go") {
			return errors.New("templ: attempt to use WriteString from a non templ file")
		}
		s, err = stringLoaderOnce().GetWatchedString(path, index, s)
		if err != nil {
			return fmt.Errorf("templ: failed to get watched string: %w", err)
		}
	}
	_, err = io.WriteString(w, s)
	return err
}

func GetDevModeTextFileName(templFileName string) string {
	if prefix, ok := strings.CutSuffix(templFileName, "_templ.go"); ok {
		templFileName = prefix + ".templ"
	}
	absFileName, err := filepath.Abs(templFileName)
	if err != nil {
		absFileName = templFileName
	}
	absFileName, err = filepath.EvalSymlinks(absFileName)
	if err != nil {
		absFileName = templFileName
	}
	absFileName = normalizePath(absFileName)

	hashedFileName := sha256.Sum256([]byte(absFileName))
	outputFileName := fmt.Sprintf("templ_%s.txt", hex.EncodeToString(hashedFileName[:]))

	root := os.TempDir()
	if os.Getenv("TEMPL_DEV_MODE_ROOT") != "" {
		root = os.Getenv("TEMPL_DEV_MODE_ROOT")
	}

	return filepath.Join(root, outputFileName)
}

// normalizePath converts Windows paths to Unix style paths.
func normalizePath(p string) string {
	p = strings.ReplaceAll(filepath.Clean(p), `\`, `/`)
	parts := strings.SplitN(p, ":", 2)
	if len(parts) == 2 && len(parts[0]) == 1 {
		drive := strings.ToLower(parts[0])
		p = "/" + drive + parts[1]
	}
	return p
}

type watchState struct {
	modTime     time.Time
	lastChecked time.Time
	strings     []string
}

// resolvedPath caches the per-caller resolution of a templ file path: the
// symlink resolution, the watch root check, and the .txt file name derived
// from them. These never change for a given caller path, so resolving them
// once keeps repeated renders from repeating the same syscalls.
type resolvedPath struct {
	txtFilePath string
	watched     bool
}

type StringLoader struct {
	watchModeRoot    string
	watchModeRootErr error
	cache            map[string]watchState
	resolved         map[string]resolvedPath
	cacheMutex       sync.Mutex
}

func NewStringLoader(devModeWatchRootPath string) (sl *StringLoader) {
	sl = &StringLoader{
		cache:    make(map[string]watchState),
		resolved: make(map[string]resolvedPath),
	}
	if devModeWatchRootPath == "" {
		return sl
	}
	resolvedRoot, err := filepath.EvalSymlinks(devModeWatchRootPath)
	if err != nil {
		sl.watchModeRootErr = fmt.Errorf("templ: failed to eval symlinks for watch mode root %q: %w", devModeWatchRootPath, err)
		return sl
	}
	sl.watchModeRoot = resolvedRoot
	return sl
}

func (sl *StringLoader) GetWatchedString(templFilePath string, index int, defaultValue string) (string, error) {
	if sl.watchModeRootErr != nil {
		return "", sl.watchModeRootErr
	}
	txtFilePath, watched, err := sl.getTxtFilePath(templFilePath)
	if err != nil {
		return "", err
	}
	// If the file is outside the watch mode root, write the string directly.
	// If watch mode root is not set, fall back to the previous behaviour to avoid breaking existing setups.
	if !watched {
		return defaultValue, nil
	}

	literals, err := sl.getWatchedStrings(txtFilePath)
	if err != nil {
		return "", fmt.Errorf("templ: failed to get watched strings for %q: %w", templFilePath, err)
	}
	if index > len(literals) {
		return "", fmt.Errorf("templ: failed to find line %d in %s", index, txtFilePath)
	}
	return strconv.Unquote(`"` + literals[index-1] + `"`)
}

// getTxtFilePath resolves the .txt file name for a templ file path, caching
// the result so symlink resolution and hashing happen once per caller path.
func (sl *StringLoader) getTxtFilePath(templFilePath string) (string, bool, error) {
	sl.cacheMutex.Lock()
	defer sl.cacheMutex.Unlock()

	if rp, ok := sl.resolved[templFilePath]; ok {
		return rp.txtFilePath, rp.watched, nil
	}

	path, err := filepath.EvalSymlinks(templFilePath)
	if err != nil {
		return "", false, fmt.Errorf("templ: failed to eval symlinks for %q: %w", templFilePath, err)
	}

	rp := resolvedPath{watched: true}
	if sl.watchModeRoot != "" && !strings.HasPrefix(path, sl.watchModeRoot) {
		rp.watched = false
	} else {
		rp.txtFilePath = GetDevModeTextFileName(path)
	}
	sl.resolved[templFilePath] = rp
	return rp.txtFilePath, rp.watched, nil
}

func (sl *StringLoader) getWatchedStrings(txtFilePath string) ([]string, error) {
	sl.cacheMutex.Lock()
	defer sl.cacheMutex.Unlock()

	state, cached := sl.cache[txtFilePath]
	if !cached {
		return sl.cacheStrings(txtFilePath)
	}

	// Throttle stat calls: only re-stat the file if it hasn't been checked in
	// the last 100ms. Throttling on the file's modTime doesn't work, because
	// the modTime is almost always older than 100ms, so the check never
	// applies and the file is stat'ed on every render.
	if time.Since(state.lastChecked) < time.Millisecond*100 {
		return state.strings, nil
	}

	info, err := os.Stat(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}
	state.lastChecked = time.Now()
	sl.cache[txtFilePath] = state

	if !info.ModTime().After(state.modTime) {
		return state.strings, nil
	}

	return sl.cacheStrings(txtFilePath)
}

func (sl *StringLoader) cacheStrings(txtFilePath string) ([]string, error) {
	txtFile, err := os.Open(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to open %s: %w", txtFilePath, err)
	}
	defer func() {
		_ = txtFile.Close()
	}()

	info, err := txtFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}

	all, err := io.ReadAll(txtFile)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to read %s: %w", txtFilePath, err)
	}

	literals := strings.Split(string(all), "\n")
	sl.cache[txtFilePath] = watchState{
		modTime:     info.ModTime(),
		lastChecked: time.Now(),
		strings:     literals,
	}

	return literals, nil
}
