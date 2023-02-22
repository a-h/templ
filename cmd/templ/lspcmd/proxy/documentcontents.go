package proxy

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	lsp "go.lsp.dev/protocol"
	"go.uber.org/zap"
)

// newDocumentContents creates a document content processing tool.
func newDocumentContents(log *zap.Logger) *DocumentContents {
	return &DocumentContents{
		m:             new(sync.Mutex),
		uriToContents: make(map[string]*Document),
		log:           log,
	}
}

type DocumentContents struct {
	m             *sync.Mutex
	uriToContents map[string]*Document
	log           *zap.Logger
}

// Set the contents of a document.
func (dc *DocumentContents) Set(uri string, d *Document) {
	dc.m.Lock()
	defer dc.m.Unlock()
	dc.uriToContents[uri] = d
}

// Get the contents of a document.
func (dc *DocumentContents) Get(uri string) (d *Document, ok bool) {
	dc.m.Lock()
	defer dc.m.Unlock()
	d, ok = dc.uriToContents[uri]
	return
}

// Delete a document from memory.
func (dc *DocumentContents) Delete(uri string) {
	dc.m.Lock()
	defer dc.m.Unlock()
	delete(dc.uriToContents, uri)
}

func (dc *DocumentContents) URIs() (uris []string) {
	dc.m.Lock()
	defer dc.m.Unlock()
	uris = make([]string, len(dc.uriToContents))
	var i int
	for k := range dc.uriToContents {
		uris[i] = k
		i++
	}
	return uris
}

// Apply changes to the document from the client, and return a list of change requests to send back to the client.
func (dc *DocumentContents) Apply(uri string, changes []lsp.TextDocumentContentChangeEvent) (d *Document, err error) {
	dc.m.Lock()
	defer dc.m.Unlock()
	var ok bool
	d, ok = dc.uriToContents[uri]
	if !ok {
		err = fmt.Errorf("document not found")
		return
	}
	for _, change := range changes {
		err = d.Overwrite(change.Range, change.Text)
		if err != nil {
			return
		}
	}
	return
}

func NewDocument(s string) *Document {
	return &Document{
		Lines: strings.Split(s, "\n"),
	}
}

type Document struct {
	Lines []string
}

func (d *Document) isEmptyRange(r lsp.Range) bool {
	return r.Start.Line == 0 && r.Start.Character == 0 &&
		r.End.Line == 0 && r.End.Character == 0
}

func (d *Document) isRangeOfDocument(r lsp.Range) bool {
	startLine, startChar := int(r.Start.Line), int(r.Start.Character)
	endLine, endChar := int(r.End.Line), int(r.End.Character)
	return startLine == 0 && startChar == 0 && endLine == len(d.Lines)-1 && endChar == len(d.Lines[len(d.Lines)-1])-1
}

func (d *Document) isOutsideDocumentRange(r lsp.Range) bool {
	startLine, startChar := int(r.Start.Line), int(r.Start.Character)
	endLine, endChar := int(r.End.Line), int(r.End.Character)
	if startLine < 0 || startChar < 0 || endChar < 0 {
		return true
	}
	startLineMaxCharIndex := len(d.Lines[startLine])
	if r.Start.Character > uint32(startLineMaxCharIndex) {
		return true
	}
	if endLine > len(d.Lines)-1 {
		return true
	}
	endLineMaxCharIndex := len(d.Lines[endLine])
	if r.End.Character > uint32(endLineMaxCharIndex) {
		return true
	}
	return false
}

func (d *Document) isWholeLineRange(r lsp.Range) bool {
	return r.Start.Character == 0 && r.End.Character == 0
}

func (d *Document) remove(i, j int) {
	d.Lines = append(d.Lines[:i], d.Lines[j:]...)
}

func (d *Document) insert(i int, withLines []string) {
	d.Lines = append(d.Lines[:i], append(withLines, d.Lines[i:]...)...)
}

var ErrOutsideDocumentRange = errors.New("range is outside of document bounds")

func (d *Document) Overwrite(r *lsp.Range, with string) error {
	if r == nil || d.isEmptyRange(*r) || d.isRangeOfDocument(*r) {
		d.Lines = strings.Split(with, "\n")
		return nil
	}
	if d.isOutsideDocumentRange(*r) {
		return ErrOutsideDocumentRange
	}
	if d.isWholeLineRange(*r) {
		d.remove(int(r.Start.Line), int(r.End.Line))
		if with != "" {
			d.insert(int(r.Start.Line), strings.Split(with, "\n"))
		}
		return nil
	}
	withLines := strings.Split(with, "\n")
	if r.Start.Character > 0 {
		prefix := d.Lines[r.Start.Line][:r.Start.Character]
		withLines[0] = prefix + withLines[0]
	}
	if r.End.Character > 0 {
		suffix := d.Lines[r.End.Line][r.End.Character:]
		withLines[len(withLines)-1] = withLines[len(withLines)-1] + suffix
	}
	d.remove(int(r.Start.Line), int(r.End.Line+1))
	d.insert(int(r.Start.Line), withLines)
	return nil
}

func (d *Document) String() string {
	return strings.Join(d.Lines, "\n")
}
