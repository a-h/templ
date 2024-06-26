package templ

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

var _ Component = JSONScriptElement{}

// JSONScript renders a JSON object inside a script element.
// e.g. <script type="application/json">{"foo":"bar"}</script>
func JSONScript(id string, data any) JSONScriptElement {
	return JSONScriptElement{
		ID:    id,
		Type:  "application/json",
		Data:  data,
		Nonce: GetNonce,
	}
}

// WithType sets the value of the type attribute of the script element.
func (j JSONScriptElement) WithType(t string) JSONScriptElement {
	j.Type = t
	return j
}

// WithNonceFromString sets the value of the nonce attribute of the script element to the given string.
func (j JSONScriptElement) WithNonceFromString(nonce string) JSONScriptElement {
	j.Nonce = func(context.Context) string {
		return nonce
	}
	return j
}

// WithNonceFrom sets the value of the nonce attribute of the script element to the value returned by the given function.
func (j JSONScriptElement) WithNonceFrom(f func(context.Context) string) JSONScriptElement {
	j.Nonce = f
	return j
}

type JSONScriptElement struct {
	// ID of the element in the DOM.
	ID string
	// Type of the script element, defaults to "application/json".
	Type string
	// Data that will be encoded as JSON.
	Data any
	// Nonce is a function that returns a CSP nonce.
	// Defaults to CSPNonceFromContext.
	// See https://content-security-policy.com/nonce for more information.
	Nonce func(ctx context.Context) string
}

func (j JSONScriptElement) Render(ctx context.Context, w io.Writer) (err error) {
	if _, err = io.WriteString(w, "<script"); err != nil {
		return err
	}
	if j.ID != "" {
		if _, err = fmt.Fprintf(w, " id=\"%s\"", EscapeString(j.ID)); err != nil {
			return err
		}
	}
	if j.Type != "" {
		if _, err = fmt.Fprintf(w, " type=\"%s\"", EscapeString(j.Type)); err != nil {
			return err
		}
	}
	if nonce := j.Nonce(ctx); nonce != "" {
		if _, err = fmt.Fprintf(w, " nonce=\"%s\"", EscapeString(nonce)); err != nil {
			return err
		}
	}
	if _, err = io.WriteString(w, ">"); err != nil {
		return err
	}
	if err = json.NewEncoder(w).Encode(j.Data); err != nil {
		return err
	}
	if _, err = io.WriteString(w, "</script>"); err != nil {
		return err
	}
	return nil
}
