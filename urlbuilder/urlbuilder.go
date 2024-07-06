package urlbuilder

import (
	"net/url"
	"strings"

	"github.com/a-h/templ"
)

// URLBuilder is a builder for constructing URLs
type URLBuilder struct {
	scheme   string
	host     string
	path     []string
	query    url.Values
	fragment string
}

// New creates a new URLBuilder with the given scheme and host
func New(scheme string, host string) *URLBuilder {
	return &URLBuilder{
		scheme: scheme,
		host:   host,
		query:  url.Values{},
	}
}

// Path adds a path segment to the URL
func (ub *URLBuilder) Path(segment string) *URLBuilder {
	ub.path = append(ub.path, segment)
	return ub
}

// Query adds a query parameter to the URL
func (ub *URLBuilder) Query(key string, value string) *URLBuilder {
	ub.query.Add(key, value)
	return ub
}

// Fragment sets the fragment (hash) part of the URL
func (ub *URLBuilder) Fragment(fragment string) *URLBuilder {
	ub.fragment = fragment
	return ub
}

// Build constructs the final URL as a SafeURL
func (ub *URLBuilder) Build() templ.SafeURL {
	var buf strings.Builder
	buf.WriteString(ub.scheme)
	buf.WriteString("://")
	buf.WriteString(ub.host)

	for _, segment := range ub.path {
		buf.WriteByte('/')
		buf.WriteString(url.PathEscape(segment))
	}

	if len(ub.query) > 0 {
		buf.WriteByte('?')
		buf.WriteString(ub.query.Encode())
	}

	if ub.fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(url.QueryEscape(ub.fragment))
	}

	return templ.URL(buf.String())
}
