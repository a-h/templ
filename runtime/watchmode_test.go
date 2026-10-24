package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetWatchedString(t *testing.T) {
	tests := []struct {
		name      string
		watchRoot string
		fileName  string
		expected  string
	}{
		{
			name:      "returns default value when file is outside watch root",
			watchRoot: "/root",
			fileName:  "/other/fileoutside_templ.go",
			expected:  "templ_file_value",
		},
		{
			name:      "uses cache when file is inside watch root",
			watchRoot: "/root",
			fileName:  "/root/fileinside_templ.go",
			expected:  "txt_file_value",
		},
		{
			name:      "uses cache when watch root is not set (legacy behaviour)",
			watchRoot: "",
			fileName:  "/root/file_templ.go",
			expected:  "txt_file_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			tmpDir := t.TempDir()

			// We have to actually make the file because GetWatchedString checks
			// the file's mod time to determine whether to use the cache or read
			// from disk.
			testFile := filepath.Join(tmpDir, tt.fileName)
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			resolvedPath, err := filepath.EvalSymlinks(testFile)
			if err != nil {
				t.Fatalf("failed to eval symlinks for test file: %v", err)
			}
			txtFile := GetDevModeTextFileName(resolvedPath)
			if err := os.WriteFile(txtFile, []byte("txt_file_value"), 0644); err != nil {
				t.Fatalf("failed to write txt file: %v", err)
			}

			watchRootPath := filepath.Join(tmpDir, tt.watchRoot)
			if err := os.MkdirAll(watchRootPath, 0755); err != nil {
				t.Fatalf("failed to create watch root directory: %v", err)
			}
			loader := NewStringLoader(watchRootPath)

			// Act.
			actual, err := loader.GetWatchedString(testFile, 1, "templ_file_value")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Assert.
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestWatchMode(t *testing.T) {
	t.Setenv("TEMPL_DEV_MODE_ROOT", "/tmp")

	t.Run("GetDevModeTextFileName respects the TEMPL_DEV_MODE_ROOT environment variable", func(t *testing.T) {
		expected := "/tmp/templ_14a26e43676c091fa17a7f4eccbbf62a44339e3cc6454b9a82c042227a21757f.txt"
		actual := GetDevModeTextFileName("test.templ")
		if actual != expected {
			t.Errorf("got %q, want %q", actual, expected)
		}
	})
	t.Run("GetDevModeTextFileName replaces _templ.go with .templ", func(t *testing.T) {
		expected := "/tmp/templ_14a26e43676c091fa17a7f4eccbbf62a44339e3cc6454b9a82c042227a21757f.txt"
		actual := GetDevModeTextFileName("test_templ.go")
		if actual != expected {
			t.Errorf("got %q, want %q", actual, expected)
		}
	})
	t.Run("GetDevModeTextFileName accepts absolute Linux paths", func(t *testing.T) {
		expected := "/tmp/templ_629591f679da14bbba764530c2965c6c8d3a8931f0ba867104c2ec441691ae22.txt"
		actual := GetDevModeTextFileName("/home/user/test.templ")
		if actual != expected {
			t.Errorf("got %q, want %q", actual, expected)
		}
	})
	t.Run("GetDevModeTextFileName accepts absolute Windows paths, which are normalized to Unix style before hashing", func(t *testing.T) {
		expected := "/tmp/templ_f0321c47222350b736aaa2d18a2b313be03da4fd4ebd80af5745434d8776376f.txt"
		actual := GetDevModeTextFileName(`C:\Windows\System32\test.templ`)
		if actual != expected {
			t.Errorf("got %q, want %q", actual, expected)
		}
	})
}

// writeWatchFiles creates a _templ.go file in dir and its dev mode .txt file, and returns
// the paths of both.
func writeWatchFiles(t testing.TB, dir, value string) (templFile, txtFile string) {
	t.Helper()
	templFile = filepath.Join(dir, "file_templ.go")
	if err := os.WriteFile(templFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write templ file: %v", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(templFile)
	if err != nil {
		t.Fatalf("failed to eval symlinks for test file: %v", err)
	}
	txtFile = GetDevModeTextFileName(resolvedPath)
	if err := os.WriteFile(txtFile, []byte(value), 0644); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}
	return templFile, txtFile
}

func TestGetWatchedStringThrottle(t *testing.T) {
	// Arrange.
	tmpDir := t.TempDir()
	t.Setenv("TEMPL_DEV_MODE_ROOT", tmpDir)
	templFile, txtFile := writeWatchFiles(t, tmpDir, "first")
	// Explicit mtimes keep the change visible on filesystems with coarse timestamps. The
	// first is well in the past, like any .txt file that wasn't just written.
	if err := os.Chtimes(txtFile, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("failed to set txt file mtime: %v", err)
	}
	loader := NewStringLoader(tmpDir)
	get := func() string {
		t.Helper()
		actual, err := loader.GetWatchedString(templFile, 1, "templ_file_value")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return actual
	}
	if actual := get(); actual != "first" {
		t.Fatalf("expected %q, got %q", "first", actual)
	}

	// Act.
	if err := os.WriteFile(txtFile, []byte("second"), 0644); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}
	if err := os.Chtimes(txtFile, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("failed to set txt file mtime: %v", err)
	}

	// Assert.
	if actual := get(); actual != "first" {
		t.Errorf("within 100ms of the last check: expected the cached %q, got %q", "first", actual)
	}
	time.Sleep(150 * time.Millisecond)
	if actual := get(); actual != "second" {
		t.Errorf("after 100ms: expected the changed %q, got %q", "second", actual)
	}
}

func TestGetWatchedStringRetriesMissingFile(t *testing.T) {
	// Arrange.
	tmpDir := t.TempDir()
	t.Setenv("TEMPL_DEV_MODE_ROOT", tmpDir)
	loader := NewStringLoader(tmpDir)
	templFile := filepath.Join(tmpDir, "file_templ.go")
	if _, err := loader.GetWatchedString(templFile, 1, "templ_file_value"); err == nil {
		t.Fatal("expected an error for a file that doesn't exist yet")
	}

	// Act.
	writeWatchFiles(t, tmpDir, "txt_file_value")
	actual, err := loader.GetWatchedString(templFile, 1, "templ_file_value")

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error once the file exists: %v", err)
	}
	if actual != "txt_file_value" {
		t.Errorf("expected %q, got %q", "txt_file_value", actual)
	}
}

func BenchmarkGetWatchedString(b *testing.B) {
	tmpDir := b.TempDir()
	b.Setenv("TEMPL_DEV_MODE_ROOT", tmpDir)
	// A path a few directories deep, like a component in a real project.
	dir := filepath.Join(tmpDir, "project", "internal", "components", "page")
	if err := os.MkdirAll(dir, 0755); err != nil {
		b.Fatalf("failed to create directory: %v", err)
	}
	templFile, _ := writeWatchFiles(b, dir, "txt_file_value")
	loader := NewStringLoader(tmpDir)
	for b.Loop() {
		if _, err := loader.GetWatchedString(templFile, 1, "templ_file_value"); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
