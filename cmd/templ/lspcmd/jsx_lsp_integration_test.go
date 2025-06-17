package lspcmd

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/a-h/templ/lsp/protocol"
)

func TestJSXAttributeLSPFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LSP integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	// Create a template with JSX components
	templContent := `package main

templ Button(text string, onClick string) {
	<button onclick={ onClick }>{ text }</button>
}

templ TestTemplate() {
	<div>
		<Button text="Click me" onClick={ fmt.Sprintf("handleClick()") } />
	</div>
}
`

	templURI := "file://" + appDir + "/jsx_test.templ"
	
	// Write the template file
	err = os.WriteFile(appDir+"/jsx_test.templ", []byte(templContent), 0644)
	if err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Open the document
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(templURI),
			LanguageID: "templ",
			Version:    1,
			Text:       templContent,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Test hover on JSX component name with timeout
	// Component "Button" is at line 8, around character 2-8
	hoverCtx, hoverCancel := context.WithTimeout(ctx, 2*time.Second)
	defer hoverCancel()
	hoverResult, err := server.Hover(hoverCtx, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 8, Character: 5}, // Middle of "Button"
		},
	})
	if err != nil {
		t.Errorf("Hover on JSX component name failed: %v", err)
	} else if hoverResult != nil {
		t.Logf("Hover on JSX component name successful: %+v", hoverResult.Contents)
	} else {
		t.Log("Hover on JSX component name returned nil (may be expected)")
	}

	// Test hover on JSX attribute expression with timeout
	// The fmt.Sprintf expression is at line 8, around character 50-75
	hoverCtx2, hoverCancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer hoverCancel2()
	hoverResult2, err := server.Hover(hoverCtx2, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 8, Character: 60}, // Inside fmt.Sprintf
		},
	})
	if err != nil {
		t.Errorf("Hover on JSX attribute expression failed: %v", err)
	} else if hoverResult2 != nil {
		t.Logf("Hover on JSX attribute expression successful: %+v", hoverResult2.Contents)
	} else {
		t.Log("Hover on JSX attribute expression returned nil (may be expected)")
	}

	// Test go-to-definition on JSX component name with timeout
	defCtx, defCancel := context.WithTimeout(ctx, 2*time.Second)
	defer defCancel()
	defResult, err := server.Definition(defCtx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 8, Character: 5}, // Middle of "Button"
		},
	})
	if err != nil {
		t.Errorf("Definition on JSX component name failed: %v", err)
	} else if defResult != nil && len(defResult) > 0 {
		t.Logf("Definition on JSX component name successful: found %d locations", len(defResult))
		for i, loc := range defResult {
			t.Logf("  Location %d: %s at %d:%d", i, loc.URI, loc.Range.Start.Line, loc.Range.Start.Character)
		}
	} else {
		t.Log("Definition on JSX component name returned no locations")
	}

	t.Log("JSX attribute LSP features test completed")
}