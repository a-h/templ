package templ

import "strings"

// FailedSanitizationURL is returned if a URL fails sanitization checks.
const FailedSanitizationURL = SafeURL("about:invalid#TemplFailedSanitizationURL")

// URL sanitizes the input string s and returns a SafeURL.
func URL(s string) SafeURL {
	if i := strings.IndexRune(s, ':'); i >= 0 && !strings.ContainsRune(s[:i], '/') {
		protocol := s[:i]
		if !strings.EqualFold(protocol, "http") && !strings.EqualFold(protocol, "https") && !strings.EqualFold(protocol, "mailto") && !strings.EqualFold(protocol, "tel") && !strings.EqualFold(protocol, "ftp") && !strings.EqualFold(protocol, "ftps") {
			return FailedSanitizationURL
		}
	}
	return SafeURL(s)
}

// SafeURL is a URL that has been sanitized.
type SafeURL string
