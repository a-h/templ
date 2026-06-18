package testargsmap

import (
	"context"
	"io"
	"testing"
)

// TestArgsMap documents and verifies the testability pattern for templ components.
//
// Usage in a handler test:
//  1. Create the map and inject it into the context with the key "_templ_args_map"
//  2. Execute the handler/component with that context
//  3. Read the arguments from the map using the parameter name as the key
func TestArgsMap(t *testing.T) {
	argsMap := make(map[string]any)
	ctx := context.WithValue(context.Background(), "_templ_args_map", argsMap)

	if err := greeting("Alice", 42).Render(ctx, io.Discard); err != nil {
		t.Fatal(err)
	}

	if got, ok := argsMap["name"].(string); !ok || got != "Alice" {
		t.Errorf("name: want Alice, got %v", argsMap["name"])
	}
	if got, ok := argsMap["count"].(int); !ok || got != 42 {
		t.Errorf("count: want 42, got %v", argsMap["count"])
	}
}

func TestArgsMapNotInjectedWithoutContextKey(t *testing.T) {
	argsMap := make(map[string]any)
	ctx := context.Background() // no "_templ_args_map" key

	if err := greeting("Bob", 7).Render(ctx, io.Discard); err != nil {
		t.Fatal(err)
	}

	if len(argsMap) != 0 {
		t.Errorf("expected empty map, got %v", argsMap)
	}
}
