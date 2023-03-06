package proxy

import (
	"fmt"
	"strings"
	"sync"

	lsp "github.com/a-h/protocol"
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
		d.Overwrite(change.Range, change.Text)
	}
	return
}

func NewDocument(log *zap.Logger, s string) *Document {
	return &Document{
		Log:   log,
		Lines: strings.Split(s, "\n"),
	}
}

type Document struct {
	Log   *zap.Logger
	Lines []string
}

func (d *Document) isEmptyRange(r lsp.Range) bool {
	return r.Start.Line == 0 && r.Start.Character == 0 &&
		r.End.Line == 0 && r.End.Character == 0
}

func (d *Document) isRangeOfDocument(r lsp.Range) bool {
	rangeStartsAtBeginningOfFile := r.Start.Line == 0 && r.Start.Character == 0
	rel, rec := int(r.End.Line), int(r.End.Character)
	del, dec := int(len(d.Lines)-1), len(d.Lines[len(d.Lines)-1])-1
	rangeEndsPastTheEndOfFile := rel > del || rel == del && rec > dec
	rangeEndsAtEndOfFile := rel == del && rec == dec
	return rangeStartsAtBeginningOfFile && (rangeEndsPastTheEndOfFile || rangeEndsAtEndOfFile)
}

func (d *Document) remove(i, j int) {
	d.Lines = append(d.Lines[:i], d.Lines[j:]...)
}

func (d *Document) insert(i int, withLines []string) {
	d.Lines = append(d.Lines[:i], append(withLines, d.Lines[i:]...)...)
}

func (d *Document) normaliseRange(r *lsp.Range) {
	if r.Start.Line > uint32(len(d.Lines))-1 {
		r.Start.Line = uint32(len(d.Lines)) - 1
	}
	if r.End.Line > uint32(len(d.Lines))-1 {
		r.End.Line = uint32(len(d.Lines)) - 1
	}
	startLine := d.Lines[r.Start.Line]
	startLineMaxCharIndex := len(startLine)
	if r.Start.Character > uint32(startLineMaxCharIndex) {
		r.Start.Character = uint32(startLineMaxCharIndex)
	}
	endLine := d.Lines[r.End.Line]
	endLineMaxCharIndex := len(endLine)
	if r.End.Character > uint32(endLineMaxCharIndex) {
		r.End.Character = uint32(endLineMaxCharIndex)
	}
}

func (d *Document) Overwrite(r *lsp.Range, with string) {
	withLines := strings.Split(with, "\n")
	if r == nil || d.isEmptyRange(*r) || len(d.Lines) == 0 {
		d.Lines = withLines
		return
	}
	d.normaliseRange(r)
	if d.isRangeOfDocument(*r) {
		d.Lines = withLines
		return
	}
	if r.Start.Character > 0 {
		prefix := d.Lines[r.Start.Line][:r.Start.Character]
		withLines[0] = prefix + withLines[0]
	}
	if r.End.Character > 0 {
		suffix := d.Lines[r.End.Line][r.End.Character:]
		withLines[len(withLines)-1] = withLines[len(withLines)-1] + suffix
	}
	if r.End.Line > r.Start.Line && r.End.Character == 0 {
		// Neovim unexpectedly adds a newline when re-inserting content (dd, followed by u for undo).
		if last := withLines[len(withLines)-1]; last == "" {
			withLines = withLines[0 : len(withLines)-1]
		}
		d.remove(int(r.Start.Line), int(r.End.Line))
	} else {
		d.remove(int(r.Start.Line), int(r.End.Line+1))
	}
	d.insert(int(r.Start.Line), withLines)
}

func (d *Document) String() string {
	return strings.Join(d.Lines, "\n")
}
