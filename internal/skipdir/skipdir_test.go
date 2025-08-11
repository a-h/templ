package skipdir

import "testing"

func TestSkipDir(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{
			name:     "current directory is not skipped",
			dir:      ".",
			expected: false,
		},
		{
			name:     "standard paths are not skipped",
			dir:      "/home/user/adrian/github.com/a-h/templ/examples",
			expected: false,
		},
		{
			name:     "vendor directories are skipped",
			dir:      "/home/user/adrian/github.com/a-h/templ/examples/vendor",
			expected: true,
		},
		{
			name:     "node_modules directories are skipped",
			dir:      "/home/user/adrian/github.com/a-h/templ/examples/node_modules",
			expected: true,
		},
		{
			name:     "dot directories are skipped",
			dir:      "/home/user/adrian/github.com/a-h/templ/examples/.git",
			expected: true,
		},
		{
			name:     "underscore directories are skipped",
			dir:      "/home/user/adrian/github.com/a-h/templ/examples/_build",
			expected: true,
		},
		{
			name:     "relative paths are normalised",
			dir:      "examples",
			expected: false,
		},
		{
			name:     "relative paths are normalised",
			dir:      "examples/vendor",
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := ShouldSkip(test.dir)
			if test.expected != actual {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
