package cors

import (
	"net/http"
	"strings"
)

type wildcard struct {
	prefix string
	suffix string
}

func (w wildcard) match(s string) bool {
	return len(s) >= len(w.prefix)+len(w.suffix) &&
		strings.HasPrefix(s, w.prefix) &&
		strings.HasSuffix(s, w.suffix)
}

// convert converts a list of string using the passed converter function
func convert(s []string, f func(string) string) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}

func first(hdrs http.Header, k string) ([]string, bool) {
	v, found := hdrs[k]
	if !found || len(v) == 0 {
		return nil, false
	}
	return v[:1], true
}
