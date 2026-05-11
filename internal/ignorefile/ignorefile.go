package ignorefile

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Patterns holds a list of glob patterns parsed from an ignore file.
type Patterns []string

// Parse reads an ignore file and returns the patterns. Returns nil and no
// error if the file does not exist.
func Parse(path string) (patterns Patterns, err error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// ShouldSkipFunc returns a function that reports whether a path should be
// skipped based on the ignore file found relative to root. If root is a file
// rather than a directory, its parent directory is used. The returned function
// accepts both absolute and relative paths; absolute paths are converted to
// relative paths before matching. If the ignore file does not exist, the
// returned function always returns false.
func ShouldSkipFunc(root, filename string) (func(string) bool, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		root = filepath.Dir(root)
	}
	patterns, err := Parse(filepath.Join(root, filename))
	if err != nil {
		return nil, err
	}
	return func(path string) bool {
		if filepath.IsAbs(path) {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return false
			}
			path = rel
		}
		return patterns.Matches(path)
	}, nil
}

// Matches reports whether the given path matches any pattern. It checks the
// full path and each directory prefix so that a pattern like "generator/test-*"
// matches "generator/test-foo/bar.templ".
func (p Patterns) Matches(path string) bool {
	if len(p) == 0 {
		return false
	}
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")
	for _, pattern := range p {
		pattern = filepath.ToSlash(pattern)
		// Check progressively longer directory prefixes, then the full path.
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			if matched, _ := filepath.Match(pattern, prefix); matched {
				return true
			}
		}
	}
	return false
}
