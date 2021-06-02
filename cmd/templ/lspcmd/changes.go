package lspcmd

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/sourcegraph/go-lsp"
	"go.uber.org/zap"
)

// newDocumentContents creates a document content processing tool.
func newDocumentContents(logger *zap.Logger) *documentContents {
	return &documentContents{
		m:             new(sync.Mutex),
		uriToContents: make(map[string][]byte),
		log:           logger,
	}
}

type documentContents struct {
	m             *sync.Mutex
	uriToContents map[string][]byte
	log           *zap.Logger
}

type documentEditor func(uri, prefix string, change lsp.TextDocumentContentChangeEvent) (requests []toClientRequest)

// Set the contents of a document.
func (fc *documentContents) Set(uri string, contents []byte) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToContents[uri] = contents
}

// Get the contents of a document.
func (fc *documentContents) Get(uri string) (contents []byte, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok = fc.uriToContents[uri]
	return
}

// Delete a document from memory.
func (fc *documentContents) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToContents, uri)
}

// Apply changes to the document from the client, and return a list of change requests to send back to the client.
func (fc *documentContents) Apply(uri string, changes []lsp.TextDocumentContentChangeEvent) (updated []byte, err error) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok := fc.uriToContents[uri]
	if !ok {
		err = fmt.Errorf("document not found")
		return
	}
	updated, err = fc.applyContentChanges(lsp.DocumentURI(uri), contents, changes)
	if err != nil {
		return
	}
	fc.uriToContents[uri] = updated
	return
}

// Contents below adapted from https://github.com/sourcegraph/go-langserver/blob/4b49d01c8a692968252730d45980091dcec7752e/langserver/fs.go#L141
// It implements the ability to react to changes on document edits.
// MIT licensed.
// applyContentChanges updates `contents` based on `changes`
func (fc *documentContents) applyContentChanges(uri lsp.DocumentURI, contents []byte, changes []lsp.TextDocumentContentChangeEvent) (c []byte, err error) {
	for _, change := range changes {
		if change.Range == nil && change.RangeLength == 0 {
			contents = []byte(change.Text) // new full content
			continue
		}
		start, ok, why := offsetForPosition(contents, change.Range.Start)
		if !ok {
			return nil, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
		}
		var end int
		if change.RangeLength != 0 {
			end = start + int(change.RangeLength)
		} else {
			// RangeLength not specified, work it out from Range.End
			end, ok, why = offsetForPosition(contents, change.Range.End)
			if !ok {
				return nil, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
			}
		}
		if start < 0 || end > len(contents) || end < start {
			return nil, fmt.Errorf("received textDocument/didChange for out of range position %q on %q", change.Range, uri)
		}
		// End of custom code.
		// Try avoid doing too many allocations, so use bytes.Buffer
		b := &bytes.Buffer{}
		b.Grow(start + len(change.Text) + len(contents) - end)
		b.Write(contents[:start])
		b.WriteString(change.Text)
		b.Write(contents[end:])
		contents = b.Bytes()
	}
	return contents, nil
}

func offsetForPosition(contents []byte, p lsp.Position) (offset int, valid bool, whyInvalid string) {
	line := 0
	col := 0
	// TODO(sqs): count chars, not bytes, per LSP. does that mean we
	// need to maintain 2 separate counters since we still need to
	// return the offset as bytes?
	for _, b := range contents {
		if line == p.Line && col == p.Character {
			return offset, true, ""
		}
		if (line == p.Line && col > p.Character) || line > p.Line {
			return 0, false, fmt.Sprintf("character %d (zero-based) is beyond line %d boundary (zero-based)", p.Character, p.Line)
		}
		offset++
		if b == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	if line == p.Line && col == p.Character {
		return offset, true, ""
	}
	if line == 0 {
		return 0, false, fmt.Sprintf("character %d (zero-based) is beyond first line boundary", p.Character)
	}
	return 0, false, fmt.Sprintf("file only has %d lines", line+1)
}

// end of content from SourceGraph.
