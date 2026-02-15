package runtime

import (
	"os"
	"path/filepath"
	"testing"
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
			os.MkdirAll(filepath.Dir(testFile), 0755)
			os.WriteFile(testFile, []byte("test"), 0644)

			resolvedPath, err := filepath.EvalSymlinks(testFile)
			if err != nil {
				t.Fatalf("failed to eval symlinks for test file: %v", err)
			}
			txtFile := GetDevModeTextFileName(resolvedPath)
			os.WriteFile(txtFile, []byte("txt_file_value"), 0644)

			watchRootPath := filepath.Join(tmpDir, tt.watchRoot)
			os.MkdirAll(watchRootPath, 0755)
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
