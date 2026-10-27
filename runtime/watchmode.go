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
	modTime time.Time
	// checkedAt is when modTime was last compared with the file on disk.
	checkedAt time.Time
	strings   []string
}

// resolvedPath is what a caller's _templ.go path resolves to. It doesn't change while the
// program runs, so it's worked out once per path instead of once per string.
type resolvedPath struct {
	outsideWatchRoot bool
	txtFilePath      string
}

// devModeCheckInterval is how long a StringLoader reuses a .txt file's strings before it
// checks the file again. It must not exceed the 100ms that templ generate waits after
// writing .txt files before it reloads the browser, or the reloaded page can render the
// previous strings.
const devModeCheckInterval = 100 * time.Millisecond

type StringLoader struct {
	// Now returns the current time. Tests replace it to control the check interval.
	Now              func() time.Time
	watchModeRoot    string
	watchModeRootErr error
	cache            map[string]watchState
	resolved         map[string]resolvedPath
	cacheMutex       sync.Mutex
}

func NewStringLoader(devModeWatchRootPath string) (sl *StringLoader) {
	sl = &StringLoader{
		Now:      time.Now,
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

	sl.cacheMutex.Lock()
	defer sl.cacheMutex.Unlock()

	rp, err := sl.resolve(templFilePath)
	if err != nil {
		return "", err
	}
	if rp.outsideWatchRoot {
		return defaultValue, nil
	}

	literals, err := sl.getWatchedStrings(rp.txtFilePath)
	if err != nil {
		return "", fmt.Errorf("templ: failed to get watched strings for %q: %w", templFilePath, err)
	}
	if index > len(literals) {
		return "", fmt.Errorf("templ: failed to find line %d in %s", index, rp.txtFilePath)
	}
	return strconv.Unquote(`"` + literals[index-1] + `"`)
}

func (sl *StringLoader) resolve(templFilePath string) (rp resolvedPath, err error) {
	rp, ok := sl.resolved[templFilePath]
	if ok {
		return rp, nil
	}
	path, err := filepath.EvalSymlinks(templFilePath)
	if err != nil {
		// Not cached, so a file that appears later still resolves.
		return rp, fmt.Errorf("templ: failed to eval symlinks for %q: %w", templFilePath, err)
	}
	// If the file is outside the watch mode root, write the string directly.
	// If watch mode root is not set, fall back to the previous behaviour to avoid breaking existing setups.
	if sl.watchModeRoot != "" && !strings.HasPrefix(path, sl.watchModeRoot) {
		rp.outsideWatchRoot = true
	} else {
		rp.txtFilePath = GetDevModeTextFileName(path)
	}
	sl.resolved[templFilePath] = rp
	return rp, nil
}

func (sl *StringLoader) getWatchedStrings(txtFilePath string) ([]string, error) {
	state, cached := sl.cache[txtFilePath]
	if !cached {
		return sl.cacheStrings(txtFilePath)
	}

	// Throttle on when the file was last checked. Its modTime is usually far in the past.
	if sl.Now().Sub(state.checkedAt) < devModeCheckInterval {
		return state.strings, nil
	}

	info, err := os.Stat(txtFilePath)
	if err != nil {
		return nil, fmt.Errorf("templ: failed to stat %s: %w", txtFilePath, err)
	}

	if !info.ModTime().After(state.modTime) {
		state.checkedAt = sl.Now()
		sl.cache[txtFilePath] = state
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
		modTime:   info.ModTime(),
		checkedAt: sl.Now(),
		strings:   literals,
	}

	return literals, nil
}
