package templ

import (
	"strings"

	"github.com/a-h/templ/templ_internal"
)

// Component is the interface that all templates implement.
type Component interface {
	templ_internal.Component
}

// ComponentFunc converts a function that matches the Component interface's
// Render method into a Component.
type ComponentFunc = templ_internal.ComponentFunc

// Hyperlink sanitization.

const failedSanitizationURL = templ_internal.SafeURL("about:invalid#TemplFailedSanitizationURL")

// URL sanitizes the input string s and returns a SafeURL.
func URL(s string) templ_internal.SafeURL {
	if i := strings.IndexRune(s, ':'); i >= 0 && !strings.ContainsRune(s[:i], '/') {
		protocol := s[:i]
		if !strings.EqualFold(protocol, "http") && !strings.EqualFold(protocol, "https") && !strings.EqualFold(protocol, "mailto") {
			return failedSanitizationURL
		}
	}
	return templ_internal.SafeURL(s)
}

// KV creates a new key/value pair from the input key and value.
func KV[TKey comparable, TValue any](key TKey, value TValue) templ_internal.KeyValue[TKey, TValue] {
	return templ_internal.KeyValue[TKey, TValue]{
		Key:   key,
		Value: value,
	}
}
