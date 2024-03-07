package processor

import (
	"os"
	"testing"
)

func TestFindTemplates(t *testing.T) {
	t.Run("returns an error if the directory does not exist", func(t *testing.T) {
		output := make(chan string)
		err := FindTemplates("nonexistent", output)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("expected os.IsNotExist(err) to be true, but got: %v", err)
		}
	})
}
