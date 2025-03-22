package runtime

import (
	"os"
	"testing"
)

func TestWatchMode(t *testing.T) {
	os.Setenv("TEMPL_DEV_MODE_ROOT", "/tmp")
	defer os.Unsetenv("TEMPL_DEV_MODE_ROOT")

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
