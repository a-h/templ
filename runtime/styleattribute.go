package runtime

import (
	"errors"
	"fmt"
	"html"
	"reflect"
	"sort"
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
	var sb strings.Builder
	for _, v := range values {
		if v == nil {
			continue
		}
		s, err := sanitizeStyleAttributeValue(v)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func sanitizeStyleAttributeValue(v any) (string, error) {
	// Process concrete types.
	switch v := v.(type) {
	case string:
		return processString(v), nil

	case templ.SafeCSS:
		return processSafeCSS(v), nil

	case map[string]string:
		return processStringMap(v), nil

	case map[string]templ.SafeCSSProperty:
		return processSafeCSSPropertyMap(v), nil

	case templ.KeyValue[string, string]:
		return processStringKV(v), nil

	case templ.KeyValue[string, bool]:
		if v.Value {
			return processString(v.Key), nil
		}
		return "", nil

	case templ.KeyValue[templ.SafeCSS, bool]:
		if v.Value {
			return processSafeCSS(v.Key), nil
		}
		return "", nil
	}

	// Fall back to reflection.

	// Handle functions first using reflection.
	if handled, result, err := handleFuncWithReflection(v); handled {
		return result, err
	}

	// Handle slices using reflection before concrete types.
	if handled, result, err := handleSliceWithReflection(v); handled {
		return result, err
	}

	return TemplUnsupportedStyleAttributeValue, nil
}

func processSafeCSS(v templ.SafeCSS) string {
	if v == "" {
		return ""
	}
	if !strings.HasSuffix(string(v), ";") {
		v += ";"
	}
	return html.EscapeString(string(v))
}

func processString(v string) string {
	if v == "" {
		return ""
	}
	sanitized := strings.TrimSpace(safehtml.SanitizeStyleValue(v))
	if !strings.HasSuffix(sanitized, ";") {
		sanitized += ";"
	}
	return html.EscapeString(sanitized)
}

var ErrInvalidStyleAttributeFunctionSignature = errors.New("invalid function signature, should be in the form func() (string, error)")

// handleFuncWithReflection handles functions using reflection.
func handleFuncWithReflection(v any) (bool, string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Func {
		return false, "", nil
	}

	t := rv.Type()
	if t.NumIn() != 0 || (t.NumOut() != 1 && t.NumOut() != 2) {
		return false, "", ErrInvalidStyleAttributeFunctionSignature
	}

	// Check the types of the return values
	if t.NumOut() == 2 {
		// Ensure the second return value is of type `error`
		secondReturnType := t.Out(1)
		if !secondReturnType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return false, "", fmt.Errorf("second return value must be of type error, got %v", secondReturnType)
		}
	}

	results := rv.Call(nil)
	var result any
	var err error

	if t.NumOut() == 2 {
		// Check if the second return value is an error
		if errVal := results[1].Interface(); errVal != nil {
			var ok bool
			err, ok = errVal.(error)
			if ok && err != nil {
				return true, "", err
			}
		}
	}

	result = results[0].Interface()

	s, err := sanitizeStyleAttributeValue(result)
	return true, s, err
}

// handleSliceWithReflection handles slices using reflection.
func handleSliceWithReflection(v any) (bool, string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return false, "", nil
	}

	var sb strings.Builder
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		s, err := sanitizeStyleAttributeValue(elem)
		if err != nil {
			return true, "", err
		}
		sb.WriteString(s)
	}
	return true, sb.String(), nil
}

// processStringMap processes a map[string]string.
func processStringMap(m map[string]string) string {
	var sb strings.Builder
	for _, name := range sortKeys(m) {
		name, value := safehtml.SanitizeCSS(name, m[name])
		sb.WriteString(name)
		sb.WriteString(":")
		sb.WriteString(value)
		sb.WriteString(";")
	}
	return html.EscapeString(sb.String())
}

// processSafeCSSPropertyMap processes a map[string]templ.SafeCSSProperty.
func processSafeCSSPropertyMap(m map[string]templ.SafeCSSProperty) string {
	var sb strings.Builder
	for _, name := range sortKeys(m) {
		sb.WriteString(safehtml.SanitizeCSSProperty(name))
		sb.WriteString(":")
		sb.WriteString(string(m[name]))
		sb.WriteString(";")
	}
	return html.EscapeString(sb.String())
}

// processStringKV processes a templ.KeyValue[string, string].
func processStringKV(kv templ.KeyValue[string, string]) string {
	name, value := safehtml.SanitizeCSS(kv.Key, kv.Value)
	return html.EscapeString(fmt.Sprintf("%s:%s;", name, value))
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

// sortKeys sorts the keys of a map.
func sortKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// TemplUnsupportedStyleAttributeValue is the default value returned for unsupported types.
var TemplUnsupportedStyleAttributeValue = "zTemplUnsupportedStyleAttributeValue:Invalid;"
