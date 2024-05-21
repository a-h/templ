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
		Data:  data,
		Nonce: GetNonce,
	}
}

func (j JSONScriptElement) WithNonceFromString(nonce string) JSONScriptElement {
	j.Nonce = func(context.Context) string {
		return nonce
	}
	return j
}

func (j JSONScriptElement) WithNonceFrom(f func(context.Context) string) JSONScriptElement {
	j.Nonce = f
	return j
}

type JSONScriptElement struct {
	// ID of the element in the DOM.
	ID string
	// Data that will be encoded as JSON.
	Data any
	// Nonce is a function that returns a CSP nonce.
	// Defaults to CSPNonceFromContext.
	// See https://content-security-policy.com/nonce for more information.
	Nonce func(ctx context.Context) string
}

func (j JSONScriptElement) Render(ctx context.Context, w io.Writer) (err error) {
	var nonceAttr string
	if nonce := j.Nonce(ctx); nonce != "" {
		nonceAttr = fmt.Sprintf(" nonce=\"%s\"", EscapeString(nonce))
	}
	if _, err = fmt.Fprintf(w, "<script id=\"%s\" type=\"application/json\"%s>", EscapeString(j.ID), nonceAttr); err != nil {
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
