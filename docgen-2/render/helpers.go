package render

import (
	"path/filepath"
	"strconv"
	"strings"
)

func titleFromPath(path string) string {
	filename, _ := baseParts(path)

	filename = strings.ReplaceAll(filename, "-", " ")

	if len(filename) > 0 {
		filename = strings.ToUpper(filename[:1]) + filename[1:]
	}

	return filename
}

func baseParts(path string) (string, int) {
	base := filepath.Base(path)
	filename := base[:len(base)-len(filepath.Ext(base))]
	prefix, suffix, hasSpace := strings.Cut(filename, "-")

	if hasSpace {
		prefix := strings.TrimPrefix(prefix, "0")
		o, err := strconv.Atoi(prefix)
		if err != nil {
			return suffix, -1
		}

		return suffix, o
	}

	return filename, -1
}
