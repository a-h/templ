package prettier

import (
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func Test(t *testing.T) {
	archive, err := txtar.ParseFile("testdata.txtar")
	if err != nil {
		t.Fatalf("failed to read testdata.txtar: %v", err)
	}
	for i := 0; i < len(archive.Files)-1; i += 2 {
		if archive.Files[i].Name != archive.Files[i+1].Name {
			t.Fatalf("test archive is not in the expected format: file pair at index %d do not match: %q vs %q", i, archive.Files[i].Name, archive.Files[i+1].Name)
		}
		t.Run(archive.Files[i].Name, func(t *testing.T) {
			inputData := archive.Files[i].Data
			expectedData := archive.Files[i+1].Data
			input := strings.TrimSpace(string(inputData))
			expected := strings.TrimSpace(string(expectedData))
			actual, err := Run(input, archive.Files[i].Name, DefaultCommand)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(actual) != expected {
				t.Errorf("Actual:\n%s\nExpected:\n%s", actual, expected)
			}
		})
	}
}
