package modcheck

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/mod/modfile"
)

func TestCheckVersion(t *testing.T) {
	const pseudoVersion = "v0.3.1101-0.20261030160437-dd6c9861fa78"
	buildInfo := func(path, version string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Path: path, Version: version}}
	}
	tests := []struct {
		name        string
		info        *debug.BuildInfo
		required    string
		wantWarning string
	}{
		{
			name:     "matching pseudo-versions despite a different embedded version",
			info:     buildInfo("github.com/a-h/templ", pseudoVersion),
			required: pseudoVersion,
		},
		{
			name:        "newer generator pseudo-version",
			info:        buildInfo("github.com/a-h/templ", pseudoVersion),
			required:    "v0.3.1101-0.20261029160437-abcdef123456",
			wantWarning: "generator " + pseudoVersion + " is newer",
		},
		{
			name:        "older generator pseudo-version",
			info:        buildInfo("github.com/a-h/templ", pseudoVersion),
			required:    "v0.3.1101-0.20261031160437-abcdef123456",
			wantWarning: "generator " + pseudoVersion + " is older",
		},
		{
			name:     "matching installed release",
			info:     buildInfo("github.com/a-h/templ", "v0.3.1100"),
			required: "v0.3.1100",
		},
		{
			name:        "newer installed release",
			info:        buildInfo("github.com/a-h/templ", "v0.3.1100"),
			required:    "v0.3.1070",
			wantWarning: "generator v0.3.1100 is newer",
		},
		{
			name:        "older installed release",
			info:        buildInfo("github.com/a-h/templ", "v0.3.1070"),
			required:    "v0.3.1100",
			wantWarning: "generator v0.3.1070 is older",
		},
		{
			name:     "local build falls back to embedded version",
			info:     buildInfo("github.com/a-h/templ", "(devel)"),
			required: templ.Version(),
		},
		{
			name:     "missing build info falls back to embedded version",
			required: templ.Version(),
		},
		{
			name:     "unrelated main module falls back to embedded version",
			info:     buildInfo("example.com/app", "v1.0.0"),
			required: templ.Version(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			mod := "module example.com/test\n\nrequire github.com/a-h/templ " + tt.required + "\n"
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0600); err != nil {
				t.Fatal(err)
			}
			err := check(dir, tt.info)
			if tt.wantWarning == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantWarning) {
				t.Fatalf("expected warning containing %q, got %v", tt.wantWarning, err)
			}
		})
	}
}

func TestWalkUp(t *testing.T) {
	t.Run("finds go.mod in a parent directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0600); err != nil {
			t.Fatal(err)
		}
		subdir := filepath.Join(dir, "child")
		if err := os.Mkdir(subdir, 0700); err != nil {
			t.Fatal(err)
		}
		got, err := WalkUp(subdir)
		if err != nil {
			t.Fatal(err)
		}
		if got != dir {
			t.Fatalf("expected module directory %q, got %q", dir, got)
		}
	})
	t.Run("returns an error when go.mod is missing", func(t *testing.T) {
		_, err := WalkUp(t.TempDir())
		if err == nil || err.Error() != "could not find go.mod file" {
			t.Fatalf("expected could not find go.mod file error, got %v", err)
		}
	})
}

func TestPatchGoVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "go 1.20",
			expected: "1.20",
		},
		{
			input:    "go 1.20.123",
			expected: "1.20.123",
		},
		{
			input:    "go 1.20.1",
			expected: "1.20.1",
		},
		{
			input:    "go 1.20rc1",
			expected: "1.20rc1",
		},
		{
			input:    "go 1.15",
			expected: "1.15",
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			input := "module github.com/a-h/templ\n\n" + string(test.input) + "\n" + "toolchain go1.27.9\n"
			mf, err := modfile.Parse("go.mod", []byte(input), nil)
			if err != nil {
				t.Fatalf("failed to parse go.mod: %v", err)
			}
			if test.expected != mf.Go.Version {
				t.Errorf("expected %q, got %q", test.expected, mf.Go.Version)
			}
		})
	}
}
