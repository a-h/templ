package runtime

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/internal/htmlattr"
)

// SanitizeAttributeName returns name if it can be written as an HTML attribute
// name, or templ.FailedSanitizationAttributeName if it can not.
func SanitizeAttributeName(name string) string {
	if !htmlattr.IsValidName(name) {
		return templ.FailedSanitizationAttributeName
	}
	return name
}

// RenderAttributes renders spread attributes on the named element. Attribute
// names that would change the structure of the element are replaced with
// templ.FailedSanitizationAttributeName. Attribute values are escaped and
// sanitized in the same way as attributes with constant names, except that
// event handler values that are not a templ.ComponentScript are replaced with
// templ.FailedSanitizationJS. The functions of templ.ComponentScript values are
// rendered to scriptWriter, which generated code writes before the opening tag
// of the element, so that the functions are defined before the element calls
// them.
func RenderAttributes(ctx context.Context, w io.Writer, scriptWriter io.Writer, elementName string, attributes templ.Attributer) (err error) {
	for _, item := range attributes.Items() {
		name := SanitizeAttributeName(item.Key)
		value := item.Value
		switch v := value.(type) {
		case templ.ComponentScript:
			if err = templ.RenderScriptItems(ctx, scriptWriter, v); err != nil {
				return err
			}
		case string, templ.SafeURL, templ.SafeCSS,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64, uintptr,
			float32, float64, complex64, complex128:
		case *string, *int, *int8, *int16, *int32, *int64,
			*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
			*float32, *float64, *complex64, *complex128:
			rv := reflect.ValueOf(v)
			if rv.IsNil() {
				continue
			}
			value = rv.Elem().Interface()
		case templ.Component:
			if htmlattr.Classify(elementName, name) != htmlattr.ContextSrcdoc {
				continue
			}
		case templ.KeyValue[string, bool]:
			if !v.Value {
				continue
			}
			value = v.Key
		case bool:
			if !v {
				continue
			}
			if _, err = io.WriteString(w, " "+templ.EscapeString(name)); err != nil {
				return err
			}
			continue
		case *bool:
			if v == nil || !*v {
				continue
			}
			if _, err = io.WriteString(w, " "+templ.EscapeString(name)); err != nil {
				return err
			}
			continue
		case templ.KeyValue[bool, bool]:
			if !v.Key || !v.Value {
				continue
			}
			if _, err = io.WriteString(w, " "+templ.EscapeString(name)); err != nil {
				return err
			}
			continue
		case func() bool:
			if !v() {
				continue
			}
			if _, err = io.WriteString(w, " "+templ.EscapeString(name)); err != nil {
				return err
			}
			continue
		default:
			continue
		}
		var resolved string
		resolved, err = resolveValue(ctx, htmlattr.Classify(elementName, name), value)
		if err != nil {
			return err
		}
		if _, err = io.WriteString(w, " "+templ.EscapeString(name)+`="`+resolved+`"`); err != nil {
			return err
		}
	}
	return nil
}

// resolveValue converts v to an HTML-escaped attribute value, and sanitizes it
// for the context.
func resolveValue(ctx context.Context, attrContext htmlattr.Context, v any) (string, error) {
	switch attrContext {
	case htmlattr.ContextEventHandler:
		script, ok := v.(templ.ComponentScript)
		if !ok {
			return templ.EscapeString(templ.FailedSanitizationJS), nil
		}
		return script.Call, nil
	case htmlattr.ContextSrcdoc:
		if c, ok := v.(templ.Component); ok {
			var sb strings.Builder
			if err := c.Render(templ.InitializeDocumentContext(ctx), &sb); err != nil {
				return "", err
			}
			return templ.EscapeString(sb.String()), nil
		}
	case htmlattr.ContextStyle:
		s, err := SanitizeStyleAttributeValues(v)
		if err != nil {
			return "", err
		}
		return templ.EscapeString(s), nil
	case htmlattr.ContextURL, htmlattr.ContextAnimationValue, htmlattr.ContextSrcset:
		if safeURL, ok := v.(templ.SafeURL); ok {
			return templ.EscapeString(safeURL), nil
		}
	default:
		if script, ok := v.(templ.ComponentScript); ok {
			return script.Call, nil
		}
	}
	s, ok := formatAttributeValue(v)
	if !ok {
		return "", fmt.Errorf("templ: unsupported attribute value type %T", v)
	}
	switch attrContext {
	case htmlattr.ContextSrcdoc:
		return templ.EscapeString(templ.EscapeString(s)), nil
	case htmlattr.ContextURL:
		return templ.EscapeString(templ.URL(s)), nil
	case htmlattr.ContextSrcset:
		return templ.EscapeString(sanitizeSrcset(s)), nil
	case htmlattr.ContextAnimationValue:
		items := strings.Split(s, ";")
		for i, item := range items {
			if templ.URL(strings.TrimSpace(item)) == templ.FailedSanitizationURL {
				items[i] = string(templ.FailedSanitizationURL)
			}
		}
		return templ.EscapeString(strings.Join(items, ";")), nil
	}
	return templ.EscapeString(s), nil
}

// formatAttributeValue formats strings, booleans, and numbers, including types
// that have one of them as the underlying type, e.g. type ID string. It returns
// false for any other type.
func formatAttributeValue(v any) (s string, ok bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case templ.ComponentScript:
		return v.Call, true
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return fmt.Sprint(v), true
	}
	return "", false
}

// sanitizeSrcset sanitizes the URL of each image candidate in a srcset. It is
// based on srcsetFilterAndEscaper and filterSrcsetElement in
// html/template/url.go, with three differences: descriptors can also contain
// '.', '+', and '-', so that density descriptors such as 1.5x are allowed; URLs
// are not normalized, because the value is HTML-escaped afterwards; and a
// rejected candidate is replaced with templ.FailedSanitizationURL.
//
// The candidates are scanned byte by byte, which is correct for UTF-8, because
// the separators of a srcset are ASCII whitespace and commas. The value is
// sanitized before it is HTML-escaped, so an entity such as &#32; in the value
// is written as text, and can't become a separator.
func sanitizeSrcset(s string) (sanitized string) {
	var sb strings.Builder
	for i, candidate := range strings.Split(s, ",") {
		if i > 0 {
			sb.WriteString(",")
		}
		writeSanitizedSrcsetCandidate(&sb, candidate)
	}
	return sb.String()
}

func writeSanitizedSrcsetCandidate(sb *strings.Builder, candidate string) {
	start := 0
	for start < len(candidate) && isHTMLSpace(candidate[start]) {
		start++
	}
	end := len(candidate)
	for i := start; i < len(candidate); i++ {
		if isHTMLSpace(candidate[i]) {
			end = i
			break
		}
	}
	if templ.URL(candidate[start:end]) != templ.FailedSanitizationURL && containsOnlySrcsetDescriptors(candidate[end:]) {
		sb.WriteString(candidate)
		return
	}
	sb.WriteString(candidate[:start])
	sb.WriteString(string(templ.FailedSanitizationURL))
}

func containsOnlySrcsetDescriptors(s string) bool {
	for i := range len(s) {
		c := s[i]
		if !isHTMLSpace(c) && !isASCIIAlnum(c) && c != '.' && c != '+' && c != '-' {
			return false
		}
	}
	return true
}

func isHTMLSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	}
	return false
}

func isASCIIAlnum(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}
