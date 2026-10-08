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
// templ.FailedSanitizationJS.
func RenderAttributes(ctx context.Context, w io.Writer, elementName string, attributes templ.Attributer) (err error) {
	for _, item := range attributes.Items() {
		name := SanitizeAttributeName(item.Key)
		value := item.Value
		switch v := value.(type) {
		case string, templ.SafeURL, templ.ComponentScript,
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
			if err := c.Render(ctx, &sb); err != nil {
				return "", err
			}
			return templ.EscapeString(sb.String()), nil
		}
	case htmlattr.ContextURL, htmlattr.ContextAnimationValue:
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
