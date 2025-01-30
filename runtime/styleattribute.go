package runtime

import (
	"errors"
	"fmt"
	"html"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/safehtml"
)

// SanitizeStyleAttributeValues renders a style attribute value.
// The supported types are:
// - string
// - templ.SafeCSS
// - map[string]string
// - map[string]templ.SafeCSSProperty
// - templ.KeyValue[string, string] - A map of key/values where the key is the CSS property name and the value is the CSS property value.
// - templ.KeyValue[string, templ.SafeCSSProperty] - A map of key/values where the key is the CSS property name and the value is the CSS property value.
// - templ.KeyValue[string, bool] - The bool determines whether the value should be included.
// - templ.KeyValue[templ.SafeCSS, bool] - The bool determines whether the value should be included.
// - func() (anyOfTheAboveTypes)
// - func() (anyOfTheAboveTypes, error)
// - []anyOfTheAboveTypes
//
// In the above, templ.SafeCSS and templ.SafeCSSProperty are types that are used to indicate that the value is safe to render as CSS without sanitization.
// All other types are sanitized before rendering.
//
// If an error is returned by any function, or a non-nil error is included in the input, the error is returned.
func SanitizeStyleAttributeValues(values ...any) (string, error) {
	if err := getJoinedErrorsFromValues(values...); err != nil {
		return "", err
	}
	sb := new(strings.Builder)
	for _, v := range values {
		if v == nil {
			continue
		}
		if err := sanitizeStyleAttributeValue(sb, v); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

func sanitizeStyleAttributeValue(sb *strings.Builder, v any) error {
	// Process concrete types.
	switch v := v.(type) {
	case string:
		return processString(sb, v)

	case templ.SafeCSS:
		return processSafeCSS(sb, v)

	case map[string]string:
		return processStringMap(sb, v)

	case map[string]templ.SafeCSSProperty:
		return processSafeCSSPropertyMap(sb, v)

	case templ.KeyValue[string, string]:
		return processStringKV(sb, v)

	case templ.KeyValue[string, bool]:
		if v.Value {
			return processString(sb, v.Key)
		}
		return nil

	case templ.KeyValue[templ.SafeCSS, bool]:
		if v.Value {
			return processSafeCSS(sb, v.Key)
		}
		return nil
	}

	// Fall back to reflection.

	// Handle functions first using reflection.
	if handled, err := handleFuncWithReflection(sb, v); handled {
		return err
	}

	// Handle slices using reflection before concrete types.
	if handled, err := handleSliceWithReflection(sb, v); handled {
		return err
	}

	_, err := sb.WriteString(TemplUnsupportedStyleAttributeValue)
	return err
}

func processSafeCSS(sb *strings.Builder, v templ.SafeCSS) error {
	if v == "" {
		return nil
	}
	sb.WriteString(html.EscapeString(string(v)))
	if !strings.HasSuffix(string(v), ";") {
		sb.WriteRune(';')
	}
	return nil
}

func processString(sb *strings.Builder, v string) error {
	if v == "" {
		return nil
	}
	sanitized := strings.TrimSpace(safehtml.SanitizeStyleValue(v))
	sb.WriteString(html.EscapeString(sanitized))
	if !strings.HasSuffix(sanitized, ";") {
		sb.WriteRune(';')
	}
	return nil
}

var ErrInvalidStyleAttributeFunctionSignature = errors.New("invalid function signature, should be in the form func() (string, error)")

// handleFuncWithReflection handles functions using reflection.
func handleFuncWithReflection(sb *strings.Builder, v any) (bool, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Func {
		return false, nil
	}

	t := rv.Type()
	if t.NumIn() != 0 || (t.NumOut() != 1 && t.NumOut() != 2) {
		return false, ErrInvalidStyleAttributeFunctionSignature
	}

	// Check the types of the return values
	if t.NumOut() == 2 {
		// Ensure the second return value is of type `error`
		secondReturnType := t.Out(1)
		if !secondReturnType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return false, fmt.Errorf("second return value must be of type error, got %v", secondReturnType)
		}
	}

	results := rv.Call(nil)

	if t.NumOut() == 2 {
		// Check if the second return value is an error
		if errVal := results[1].Interface(); errVal != nil {
			if err, ok := errVal.(error); ok && err != nil {
				return true, err
			}
		}
	}

	return true, sanitizeStyleAttributeValue(sb, results[0].Interface())
}

// handleSliceWithReflection handles slices using reflection.
func handleSliceWithReflection(sb *strings.Builder, v any) (bool, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return false, nil
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if err := sanitizeStyleAttributeValue(sb, elem); err != nil {
			return true, err
		}
	}
	return true, nil
}

// processStringMap processes a map[string]string.
func processStringMap(sb *strings.Builder, m map[string]string) error {
	for _, name := range slices.Sorted(maps.Keys(m)) {
		name, value := safehtml.SanitizeCSS(name, m[name])
		sb.WriteString(html.EscapeString(name))
		sb.WriteRune(':')
		sb.WriteString(html.EscapeString(value))
		sb.WriteRune(';')
	}
	return nil
}

// processSafeCSSPropertyMap processes a map[string]templ.SafeCSSProperty.
func processSafeCSSPropertyMap(sb *strings.Builder, m map[string]templ.SafeCSSProperty) error {
	for _, name := range slices.Sorted(maps.Keys(m)) {
		sb.WriteString(html.EscapeString(safehtml.SanitizeCSSProperty(name)))
		sb.WriteRune(':')
		sb.WriteString(html.EscapeString(string(m[name])))
		sb.WriteRune(';')
	}
	return nil
}

// processStringKV processes a templ.KeyValue[string, string].
func processStringKV(sb *strings.Builder, kv templ.KeyValue[string, string]) error {
	name, value := safehtml.SanitizeCSS(kv.Key, kv.Value)
	sb.WriteString(html.EscapeString(name))
	sb.WriteRune(':')
	sb.WriteString(html.EscapeString(value))
	sb.WriteRune(';')
	return nil
}

// getJoinedErrorsFromValues collects and joins errors from the input values.
func getJoinedErrorsFromValues(values ...any) error {
	var errs []error
	for _, v := range values {
		if err, ok := v.(error); ok {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// TemplUnsupportedStyleAttributeValue is the default value returned for unsupported types.
var TemplUnsupportedStyleAttributeValue = "zTemplUnsupportedStyleAttributeValue:Invalid;"
