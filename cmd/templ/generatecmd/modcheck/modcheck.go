package modcheck

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

// WalkUp the directory tree, starting at dir, until we find a directory containing
// a go.mod file.
func WalkUp(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	var modFile string
	for {
		modFile = filepath.Join(dir, "go.mod")
		_, err := os.Stat(modFile)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to stat go.mod file: %w", err)
		}
		if os.IsNotExist(err) {
			// Move up.
			prev := dir
			dir = filepath.Dir(dir)
			if dir == prev {
				break
			}
			continue
		}
		break
	}

	// No file found.
	if modFile == "" {
		return dir, fmt.Errorf("could not find go.mod file")
	}
	return dir, nil
}

func Check(dir string) error {
	dir, err := WalkUp(dir)
	if err != nil {
		return err
	}

	// Found a go.mod file.
	// Read it and find the templ version.
	modFile := filepath.Join(dir, "go.mod")
	m, err := os.ReadFile(modFile)
	if err != nil {
		return fmt.Errorf("failed to read go.mod file: %w", err)
	}

	mf, err := modfile.Parse(modFile, m, nil)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod file: %w", err)
	}
	if mf.Module.Mod.Path == "github.com/a-h/templ" {
		// The go.mod file is for templ itself.
		return nil
	}
	for _, r := range mf.Require {
		if r.Mod.Path == "github.com/a-h/templ" {
			cmp := semver.Compare(r.Mod.Version, templ.Version())
			if cmp < 0 {
				return fmt.Errorf("generator %v is newer than templ version %v found in go.mod file, consider running `go get -u github.com/a-h/templ` to upgrade", templ.Version(), r.Mod.Version)
			}
			if cmp > 0 {
				return fmt.Errorf("generator %v is older than templ version %v found in go.mod file, consider upgrading templ CLI", templ.Version(), r.Mod.Version)
			}
			return nil
		}
	}
	return fmt.Errorf("templ not found in go.mod file, run `go get github.com/a-h/templ` to install it")
}
