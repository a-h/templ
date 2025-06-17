package lspcmd

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"
)

func TestJSXLSP(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	// Create a templ file with JSX syntax
	jsxTemplContent := `package main

import "fmt"

templ Button(title string) {
	<button>{ title }</button>
}

templ JSXPage() {
	<!DOCTYPE html>
	<html>
		<body>
			<h1>JSX Test</h1>
			<Button title="Click me" />
			<Button title={ fmt.Sprintf("Dynamic %s", "button") } />
		</body>
	</html>
}
`

	jsxFileName := appDir + "/jsx_test.templ"
	err = os.WriteFile(jsxFileName, []byte(jsxTemplContent), 0644)
	if err != nil {
		t.Fatalf("failed to create JSX test file: %v", err)
	}

	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + jsxFileName),
			LanguageID: "templ",
			Version:    1,
			Text:       jsxTemplContent,
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}

	log.Info("Testing JSX hover")

	// Test hover on JSX component
	tests := []struct {
		name        string
		line        int
		replacement string
		cursor      string
		assert      func(t *testing.T, hover *protocol.Hover) (msg string, ok bool)
	}{
		{
			name:        "hover-on-jsx-component",
			line:        13,
			replacement: `			<Button title="Click me" />`,
			cursor:      `			 ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				// We just want to make sure this doesn't panic
				log.Info("Hover result", slog.Any("hover", actual))
				return "", true
			},
		},
		{
			name:        "hover-on-jsx-attribute", 
			line:        13,
			replacement: `			<Button title="Click me" />`,
			cursor:      `			        ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				// We just want to make sure this doesn't panic
				log.Info("Hover result", slog.Any("hover", actual))
				return "", true
			},
		},
		{
			name:        "hover-on-jsx-expression",
			line:        14,
			replacement: `			<Button title={ fmt.Sprintf("Dynamic %s", "button") } />`,
			cursor:      `			                ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				// We just want to make sure this doesn't panic
				log.Info("Hover result", slog.Any("hover", actual))
				return "", true
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset file content
			err = server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + jsxFileName),
					},
					Version: 2,
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					{
						Range: nil,
						Text:  jsxTemplContent,
					},
				},
			})
			if err != nil {
				t.Errorf("failed to change file: %v", err)
				return
			}

			// Give some time for processing
			time.Sleep(time.Millisecond * 100)

			lspCharIndex, err := runeIndexToUTF8ByteIndex(test.replacement, len(test.cursor)-1)
			if err != nil {
				t.Error(err)
				return
			}

			actual, err := server.Hover(ctx, &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + jsxFileName),
					},
					Position: protocol.Position{
						Line:      uint32(test.line - 1),
						Character: lspCharIndex,
					},
				},
			})
			if err != nil {
				t.Errorf("failed to hover: %v", err)
				return
			}

			msg, ok := test.assert(t, actual)
			if !ok {
				t.Error(msg)
			}
		})
	}
}

func TestJSXSourceMapPanic(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	// Create a JSX template that parses successfully 
	jsxTemplContent := `package main

import "fmt"

templ Button(title string) {
	<button>{ title }</button>
}

templ JSXComplexPage() {
	<!DOCTYPE html>
	<html>
		<body>
			<h1>JSX Complex Test</h1>
			<Button title="Simple button" />
			<Button title={ fmt.Sprintf("Dynamic %s", "content") } />
			<div>
				<Button title="Nested button" />
			</div>
		</body>
	</html>
}
`

	jsxFileName := appDir + "/jsx_panic_test.templ"
	err = os.WriteFile(jsxFileName, []byte(jsxTemplContent), 0644)
	if err != nil {
		t.Fatalf("failed to create JSX test file: %v", err)
	}

	// This mimics what happens when LSP processes a file
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + jsxFileName),
			LanguageID: "templ",
			Version:    1,
			Text:       jsxTemplContent,
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}

	// Give some time for background processing that might trigger the panic
	time.Sleep(time.Millisecond * 500)

	// Try various LSP operations that might trigger the panic
	tests := []struct {
		name      string
		operation func() error
	}{
		{
			name: "hover-on-jsx-button",
			operation: func() error {
				_, err := server.Hover(ctx, &protocol.HoverParams{
					TextDocumentPositionParams: protocol.TextDocumentPositionParams{
						TextDocument: protocol.TextDocumentIdentifier{
							URI: uri.URI("file://" + jsxFileName),
						},
						Position: protocol.Position{
							Line:      uint32(9), // <Button title="Simple button" />
							Character: uint32(4),  // On the "B" of Button
						},
					},
				})
				return err
			},
		},
		{
			name: "completion-in-jsx-expression",
			operation: func() error {
				_, err := server.Completion(ctx, &protocol.CompletionParams{
					TextDocumentPositionParams: protocol.TextDocumentPositionParams{
						TextDocument: protocol.TextDocumentIdentifier{
							URI: uri.URI("file://" + jsxFileName),
						},
						Position: protocol.Position{
							Line:      uint32(10), // <Button title={ fmt.Sprintf(...) } />
							Character: uint32(20), // Inside the expression
						},
					},
				})
				return err
			},
		},
		{
			name: "document-symbol",
			operation: func() error {
				_, err := server.DocumentSymbol(ctx, &protocol.DocumentSymbolParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + jsxFileName),
					},
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			log.Info("Running operation", slog.String("operation", test.name))
			
			// The main goal is to ensure these don't panic
			err := test.operation()
			if err != nil {
				t.Logf("Operation %s returned error (this is ok, we're testing for panics): %v", test.name, err)
			}
			
			log.Info("Operation completed without panic", slog.String("operation", test.name))
		})
	}
}

func TestJSXPublishDiagnosticsNilPanic(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	// Create a JSX template with syntax errors that will trigger diagnostics
	jsxTemplContent := `package main

import "fmt"

templ Button(title string) {
	<button>{ title }</button>
}

templ JSXPageWithErrors() {
	<!DOCTYPE html>
	<html>
		<body>
			<h1>JSX Test</h1>
			<Button title="Simple button" />
			<Button title={ fmt.Sprintf("Dynamic %s", "content") } />
			<div>
				<Button title="Nested button" />
				{ undefinedVariable }
			</div>
		</body>
	</html>
}
`

	jsxFileName := appDir + "/jsx_diagnostics_test.templ"
	err = os.WriteFile(jsxFileName, []byte(jsxTemplContent), 0644)
	if err != nil {
		t.Fatalf("failed to create JSX test file: %v", err)
	}

	// Open the file - this should trigger generation and source map caching
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + jsxFileName),
			LanguageID: "templ",
			Version:    1,
			Text:       jsxTemplContent,
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}

	// Wait for background processing that should trigger PublishDiagnostics
	// This is where the nil pointer panic should occur if sourceMap is nil
	time.Sleep(time.Millisecond * 2000)

	// Create a second file that has compilation errors in the generated Go code
	// This will more likely trigger the PublishDiagnostics path that causes the panic
	jsxTemplWithGoErrors := `package main

import "fmt"
import "nonexistent/package"

templ ButtonWithErrors(title string) {
	<button>{ title }</button>
}

templ JSXPageWithGoErrors() {
	<!DOCTYPE html>
	<html>
		<body>
			<h1>JSX Test</h1>
			<ButtonWithErrors title="Simple button" />
			<ButtonWithErrors title={ fmt.Sprintf("Dynamic %s", nonexistent.Func()) } />
		</body>
	</html>
}
`

	jsxFileNameErrors := appDir + "/jsx_go_errors_test.templ"
	err = os.WriteFile(jsxFileNameErrors, []byte(jsxTemplWithGoErrors), 0644)
	if err != nil {
		t.Fatalf("failed to create JSX test file with errors: %v", err)
	}

	// Open the file with Go compilation errors
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + jsxFileNameErrors),
			LanguageID: "templ",
			Version:    1,
			Text:       jsxTemplWithGoErrors,
		},
	})
	if err != nil {
		t.Errorf("failed to register open file with errors: %v", err)
		return
	}

	// Wait longer for background Go compilation and diagnostics publishing
	// The panic should occur here when PublishDiagnostics tries to map Go errors back to templ
	time.Sleep(time.Millisecond * 5000)

	// If we get here without panic, the fix worked
	log.Info("PublishDiagnostics test completed without panic")
}

func TestJSXCompletion(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	// Create a templ file with JSX syntax
	jsxTemplContent := `package main

import "fmt"

templ Button(title string) {
	<button>{ title }</button>
}

templ JSXPage() {
	<!DOCTYPE html>
	<html>
		<body>
			<h1>JSX Test</h1>
			<Button title="Click me" />
		</body>
	</html>
}
`

	jsxFileName := appDir + "/jsx_completion_test.templ"
	err = os.WriteFile(jsxFileName, []byte(jsxTemplContent), 0644)
	if err != nil {
		t.Fatalf("failed to create JSX test file: %v", err)
	}

	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + jsxFileName),
			LanguageID: "templ",
			Version:    1,
			Text:       jsxTemplContent,
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}

	log.Info("Testing JSX completion")

	// Test completion inside JSX expression
	tests := []struct {
		name        string
		line        int
		replacement string
		cursor      string
		assert      func(t *testing.T, cl *protocol.CompletionList) (msg string, ok bool)
	}{
		{
			name:        "completion-in-jsx-expression",
			line:        14,
			replacement: `			<Button title={ fm`,
			cursor:      `			               ^`,
			assert: func(t *testing.T, actual *protocol.CompletionList) (msg string, ok bool) {
				// We just want to make sure this doesn't panic
				log.Info("Completion result", slog.Any("completion", actual))
				return "", true
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Edit the file for completion test
			updated := jsxTemplContent
			lines := []string{
				`package main`,
				``,
				`import "fmt"`,
				``,
				`templ Button(title string) {`,
				`	<button>{ title }</button>`,
				`}`,
				``,
				`templ JSXPage() {`,
				`	<!DOCTYPE html>`,
				`	<html>`,
				`		<body>`,
				`			<h1>JSX Test</h1>`,
				test.replacement,
				`		</body>`,
				`	</html>`,
				`}`,
			}
			updated = ""
			for i, line := range lines {
				if i > 0 {
					updated += "\n"
				}
				updated += line
			}

			err = server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + jsxFileName),
					},
					Version: 2,
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					{
						Range: nil,
						Text:  updated,
					},
				},
			})
			if err != nil {
				t.Errorf("failed to change file: %v", err)
				return
			}

			// Give some time for processing
			time.Sleep(time.Millisecond * 100)

			actual, err := server.Completion(ctx, &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + jsxFileName),
					},
					Position: protocol.Position{
						Line:      uint32(test.line - 1),
						Character: uint32(len(test.cursor) - 1),
					},
				},
			})
			if err != nil {
				t.Errorf("failed to get completion: %v", err)
				return
			}

			msg, ok := test.assert(t, actual)
			if !ok {
				t.Error(msg)
			}
		})
	}
}