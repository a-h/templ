package runtime

import (
	"context"
	"strings"
	"testing"
)

func TestGeneratedTemplate(t *testing.T) {
	f := func(input GeneratedComponentInput) error {
		_, err := input.Writer.Write([]byte("Hello, World!"))
		return err
	}
	sb := new(strings.Builder)
	err := GeneratedTemplate(f).Render(context.Background(), sb)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sb.String() != "Hello, World!" {
		t.Errorf("expected \"Hello, World!\", got %q", sb.String())
	}
}
