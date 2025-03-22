package skipdir

import (
	"path/filepath"
	"strings"
)

func ShouldSkip(path string) (skip bool) {
	if path == "." {
		return false
	}
	_, name := filepath.Split(path)
	if name == "vendor" || name == "node_modules" {
		return true
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return true
	}
	return false
}
