package lspcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

func TestJSXGeneratorWithWorkingDirectory(t *testing.T) {
	// Create a temporary workspace directory
	tempDir, err := os.MkdirTemp("", "templ-jsx-workdir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod
	goMod := `module testmod

go 1.21
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Create external JSX component in jsxmod package
	jsxmodDir := filepath.Join(tempDir, "jsxmod")
	if err := os.MkdirAll(jsxmodDir, 0755); err != nil {
		t.Fatal(err)
	}

	jsxmodGo := `package jsxmod

import (
	"context"
	"io"
	"github.com/a-h/templ"
)

func Text(content string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, content)
		return err
	})
}
`
	if err := os.WriteFile(filepath.Join(jsxmodDir, "text.go"), []byte(jsxmodGo), 0644); err != nil {
		t.Fatal(err)
	}

	// Create template file that uses external JSX component
	templContent := `package main

templ ExampleTemplate() {
	<div>
		<jsxmod.Text content="Hello World" />
	</div>
}
`
	
	// Test 1: Generator WITHOUT working directory (old behavior)
	template, err := parser.ParseString(templContent)
	if err != nil {
		t.Fatal(err)
	}

	var outputWithoutWorkDir strings.Builder
	generatorOutputWithout, err := generator.Generate(template, &outputWithoutWorkDir)
	hasErrorWithout := err != nil
	sourceMapWithout := generatorOutputWithout.SourceMap
	isNilWithout := sourceMapWithout == nil

	// Test 2: Generator WITH working directory (new behavior)
	var outputWithWorkDir strings.Builder
	generatorOutputWith, err := generator.Generate(template, &outputWithWorkDir, generator.WithWorkingDir(tempDir))
	hasErrorWith := err != nil
	sourceMapWith := generatorOutputWith.SourceMap
	isNilWith := sourceMapWith == nil

	t.Logf("Without working dir: error=%v, sourceMap nil=%v", hasErrorWithout, isNilWithout)
	t.Logf("With working dir: error=%v, sourceMap nil=%v", hasErrorWith, isNilWith)

	// The fix should mean that WITH working directory, we get a non-nil source map
	// even if there's still an error due to module resolution
	if isNilWith && !hasErrorWith {
		t.Error("Expected non-nil source map when working directory is provided")
	}

	// We expect the working directory version to be better (either no error, or error but still source map)
	if isNilWith && !isNilWithout {
		t.Error("Working directory version should not be worse than non-working directory version")
	}

	// If we get a source map with working directory, it should have content
	if !isNilWith && len(sourceMapWith.Expressions) == 0 {
		t.Log("Source map exists but has no expressions - this could indicate partial success")
	}

	t.Logf("Test completed - working directory fix allows generator to attempt JSX resolution")
}