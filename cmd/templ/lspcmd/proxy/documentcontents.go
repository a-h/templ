package proxy

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	lsp "github.com/a-h/templ/lsp/protocol"
)

// newDocumentContents creates a document content processing tool.
func newDocumentContents(log *slog.Logger) *DocumentContents {
	return &DocumentContents{
		m:             new(sync.Mutex),
		uriToContents: make(map[string]*Document),
		log:           log,
	}
}

type DocumentContents struct {
	m             *sync.Mutex
	uriToContents map[string]*Document
	log           *slog.Logger
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
		d.Apply(change.Range, change.Text)
	}
	return
}

func NewDocument(log *slog.Logger, s string) *Document {
	return &Document{
		Log:   log,
		Lines: strings.Split(s, "\n"),
	}
}

type Document struct {
	Log   *slog.Logger
	Lines []string
}

func (d *Document) LineLengths() (lens []int) {
	lens = make([]int, len(d.Lines))
	for i, l := range d.Lines {
		lens[i] = len(l)
	}
	return
}

func (d *Document) Len() (line, col int) {
	line = len(d.Lines)
	col = len(d.Lines[len(d.Lines)-1])
	return
}

func (d *Document) Overwrite(fromLine, fromCol, toLine, toCol int, lines []string) {
	suffix := d.Lines[toLine][toCol:]
	toLen := d.LineLengths()[toLine]
	d.Delete(fromLine, fromCol, toLine, toLen)
	lines[len(lines)-1] = lines[len(lines)-1] + suffix
	d.Insert(fromLine, fromCol, lines)
}

func (d *Document) Insert(line, col int, lines []string) {
	prefix := d.Lines[line][:col]
	suffix := d.Lines[line][col:]
	lines[0] = prefix + lines[0]
	d.Lines[line] = lines[0]

	if len(lines) > 1 {
		d.InsertLines(line+1, lines[1:])
	}

	d.Lines[line+len(lines)-1] = lines[len(lines)-1] + suffix
}

func (d *Document) InsertLines(i int, withLines []string) {
	d.Lines = append(d.Lines[:i], append(withLines, d.Lines[i:]...)...)
}

func (d *Document) Delete(fromLine, fromCol, toLine, toCol int) {
	prefix := d.Lines[fromLine][:fromCol]
	suffix := d.Lines[toLine][toCol:]

	// Delete intermediate lines.
	deleteFrom := fromLine
	deleteTo := fromLine + (toLine - fromLine)
	d.DeleteLines(deleteFrom, deleteTo)

	// Merge the contents of the final line.
	d.Lines[fromLine] = prefix + suffix
}

func (d *Document) DeleteLines(i, j int) {
	d.Lines = append(d.Lines[:i], d.Lines[j:]...)
}

func (d *Document) String() string {
	return strings.Join(d.Lines, "\n")
}

func (d *Document) Replace(with string) {
	d.Lines = strings.Split(with, "\n")
}

func (d *Document) Apply(r *lsp.Range, with string) {
	withLines := strings.Split(with, "\n")
	d.normalize(r)
	if d.isWholeDocument(r) {
		d.Lines = withLines
		return
	}
	if d.isInsert(r, with) {
		d.Insert(int(r.Start.Line), int(r.Start.Character), withLines)
		return
	}
	if d.isDelete(r, with) {
		d.Delete(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character))
		return
	}
	if d.isOverwrite(r, with) {
		d.Overwrite(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character), withLines)
	}
}

func (d *Document) normalize(r *lsp.Range) {
	if r == nil {
		return
	}
	lens := d.LineLengths()
	if r.Start.Line >= uint32(len(lens)) {
		r.Start.Line = uint32(len(lens) - 1)
		r.Start.Character = uint32(lens[r.Start.Line])
	}
	if r.Start.Character > uint32(lens[r.Start.Line]) {
		r.Start.Character = uint32(lens[r.Start.Line])
	}
	if r.End.Line >= uint32(len(lens)) {
		r.End.Line = uint32(len(lens) - 1)
		r.End.Character = uint32(lens[r.End.Line])
	}
	if r.End.Character > uint32(lens[r.End.Line]) {
		r.End.Character = uint32(lens[r.End.Line])
	}
}

func (d *Document) isOverwrite(r *lsp.Range, with string) bool {
	return (r.End.Line != r.Start.Line || r.Start.Character != r.End.Character) && with != ""
}

func (d *Document) isInsert(r *lsp.Range, with string) bool {
	return r.End.Line == r.Start.Line && r.Start.Character == r.End.Character && with != ""
}

func (d *Document) isDelete(r *lsp.Range, with string) bool {
	return (r.End.Line != r.Start.Line || r.Start.Character != r.End.Character) && with == ""
}

func (d *Document) isWholeDocument(r *lsp.Range) bool {
	if r == nil {
		return true
	}
	if r.Start.Line != 0 || r.Start.Character != 0 {
		return false
	}
	l, c := d.Len()
	return r.End.Line == uint32(l) || r.End.Character == uint32(c)
}
