package predeclared

import "strconv"

// FormatString formats a string, int, int64, or bool in the same way as
// fmt.Sprint, without the cost of fmt. It returns false for any other type,
// including named types, which fmt.Sprint formats with their String method.
func FormatString(v any) (s string, ok bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case int:
		return strconv.Itoa(v), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case bool:
		return strconv.FormatBool(v), true
	}
	return "", false
}
