package templ

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ImportMap represents an import map JSON object.
type ImportMap struct {
	Imports   map[string]string            `json:"imports"`
	Scopes    map[string]map[string]string `json:"scopes,omitempty"`
	Integrity map[string]string            `json:"integrity,omitempty"`
}

// NewImportMapScriptElement creates a new ImportMapScriptElement.
// The data must be a valid JSON string.
func NewImportMapScriptElement(id string, data string) ImportMapScriptElement {
	im := ImportMap{}
	if err := json.Unmarshal([]byte(data), &im); err != nil {
		panic(fmt.Sprintf("invalid importmap JSON data: %v", err))
	}
	// We re-encode the JSON data to dismiss any potential spurious data.
	// The jspm cli install command adds a "env" key to the import map JSON data.
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "")
	enc.SetEscapeHTML(false)
	err := enc.Encode(im)
	if err != nil {
		panic(fmt.Sprintf("failed to re-encode importmap JSON data: %v", err))
	}
	return ImportMapScriptElement{
		id:   id,
		data: buf.String(),
	}
}

var _ Component = ImportMapScriptElement{}

type ImportMapScriptElement struct {
	id   string
	data string
}

func (j ImportMapScriptElement) Render(ctx context.Context, w io.Writer) (err error) {
	if _, err = io.WriteString(w, "<script type=\"importmap\""); err != nil {
		return err
	}
	if j.id != "" {
		if _, err = fmt.Fprintf(w, " id=\"%s\"", EscapeString(j.id)); err != nil {
			return err
		}
	}
	if _, err = io.WriteString(w, ">"+j.data+"</script>"); err != nil {
		return err
	}
	return nil
}
