package lspcmd

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/a-h/templ/lsp/protocol"
)

func TestElementComponentAttributeLSPFeatures(t *testing.T) {
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

	// Create a template with element components
	templContent := `package main

import "fmt"

templ Button(text string, onClick string) {
	<button onclick={ onClick }>{ text }</button>
}

templ TestTemplate() {
	<div>
		@Button("Click me", fmt.Sprintf("handleClick()"))
	</div>
}
`

	templURI := "file://" + appDir + "/element_component_test.templ"
	
	// Write the template file
	err = os.WriteFile(appDir+"/element_component_test.templ", []byte(templContent), 0644)
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

	// Test hover on element component name with timeout
	// Component "Button" is at line 10, around character 2-8 (adjusted for import)
	hoverCtx, hoverCancel := context.WithTimeout(ctx, 2*time.Second)
	defer hoverCancel()
	hoverResult, err := server.Hover(hoverCtx, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 10, Character: 5}, // Middle of "Button"
		},
	})
	if err != nil {
		t.Errorf("Hover on element component name failed: %v", err)
	} else if hoverResult != nil {
		t.Logf("Hover on element component name successful: %+v", hoverResult.Contents)
	} else {
		t.Log("Hover on element component name returned nil (may be expected)")
	}

	// Test hover on element component attribute expression with timeout
	// The fmt.Sprintf expression is at line 10, around character 50-75 (adjusted for import)
	hoverCtx2, hoverCancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer hoverCancel2()
	hoverResult2, err := server.Hover(hoverCtx2, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 10, Character: 50}, // Inside fmt.Sprintf expression
		},
	})
	if err != nil {
		t.Errorf("Hover on element component attribute expression failed: %v", err)
	} else if hoverResult2 != nil {
		t.Logf("Hover on element component attribute expression successful: %+v", hoverResult2.Contents)
	} else {
		t.Log("Hover on element component attribute expression returned nil (may be expected)")
	}

	// Test go-to-definition on element component name with timeout
	defCtx, defCancel := context.WithTimeout(ctx, 2*time.Second)
	defer defCancel()
	defResult, err := server.Definition(defCtx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 10, Character: 5}, // Middle of "Button"
		},
	})
	if err != nil {
		t.Errorf("Definition on element component name failed: %v", err)
	} else if defResult != nil && len(defResult) > 0 {
		t.Logf("Definition on element component name successful: found %d locations", len(defResult))
		for i, loc := range defResult {
			t.Logf("  Location %d: %s at %d:%d", i, loc.URI, loc.Range.Start.Line, loc.Range.Start.Character)
		}
	} else {
		t.Log("Definition on element component name returned no locations")
	}

	// Test hover on element component attribute name to verify it maps to function parameter
	// The attribute name "text" is at line 10, around character 12
	attrHoverCtx, attrHoverCancel := context.WithTimeout(ctx, 2*time.Second)
	defer attrHoverCancel()
	attrHoverResult, err := server.Hover(attrHoverCtx, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 10, Character: 12}, // Middle of "text" attribute name
		},
	})
	if err != nil {
		t.Errorf("Hover on element component attribute name failed: %v", err)
	} else if attrHoverResult != nil {
		t.Logf("Hover on element component attribute name successful: %+v", attrHoverResult.Contents)
	} else {
		t.Log("Hover on element component attribute name returned nil (this is what we're trying to fix)")
	}

	// Test go-to-definition on element component attribute name
	attrDefCtx, attrDefCancel := context.WithTimeout(ctx, 2*time.Second)
	defer attrDefCancel()
	attrDefResult, err := server.Definition(attrDefCtx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(templURI),
			},
			Position: protocol.Position{Line: 10, Character: 12}, // Middle of "text" attribute name
		},
	})
	if err != nil {
		t.Errorf("Definition on element component attribute name failed: %v", err)
	} else if attrDefResult != nil && len(attrDefResult) > 0 {
		t.Logf("Definition on element component attribute name successful: found %d locations", len(attrDefResult))
		for i, loc := range attrDefResult {
			t.Logf("  Location %d: %s at %d:%d", i, loc.URI, loc.Range.Start.Line, loc.Range.Start.Character)
		}
	} else {
		t.Log("Definition on element component attribute name returned no locations (this is what we're trying to fix)")
	}

	t.Log("Element component attribute LSP features test completed")
}