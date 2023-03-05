package httpdebug

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/lspcmd/proxy"
	"github.com/a-h/templ/cmd/templ/visualize"
)

func NewHandler(s *proxy.Server) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/templ", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		c, ok := s.TemplSource.Get(uri)
		if !ok {
			Error(w, "uri not found", http.StatusNotFound)
			return
		}
		io.WriteString(w, c.String())
	})
	m.HandleFunc("/sourcemap", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		sm, ok := s.SourceMapCache.Get(uri)
		if !ok {
			Error(w, "uri not found", http.StatusNotFound)
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(sm.SourceLinesToTarget)
	})
	m.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		c, ok := s.GoSource[uri]
		if !ok {
			Error(w, "uri not found", http.StatusNotFound)
			return
		}
		io.WriteString(w, c)
	})
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		if uri == "" {
			// List all URIs.
			list(s.TemplSource.URIs()).Render(r.Context(), w)
			return
		}
		// Assume we've got a URI.
		templSource, ok := s.TemplSource.Get(uri)
		if !ok {
			if !ok {
				Error(w, "uri not found in document contents", http.StatusNotFound)
				return
			}
		}
		goSource, ok := s.GoSource[uri]
		if !ok {
			if !ok {
				Error(w, "uri not found in document contents", http.StatusNotFound)
				return
			}
		}
		sm, ok := s.SourceMapCache.Get(uri)
		if !ok {
			Error(w, "uri not found", http.StatusNotFound)
			return
		}
		visualize.HTML(uri, templSource.String(), goSource, sm).Render(r.Context(), w)
	})
	return m
}

func getMapURL(uri string) templ.SafeURL {
	return withQuery("/", uri)
}

func getSourceMapURL(uri string) templ.SafeURL {
	return withQuery("/sourcemap", uri)
}

func getTemplURL(uri string) templ.SafeURL {
	return withQuery("/templ", uri)
}

func getGoURL(uri string) templ.SafeURL {
	return withQuery("/go", uri)
}

func withQuery(path, uri string) templ.SafeURL {
	q := make(url.Values)
	q.Set("uri", uri)
	u := &url.URL{
		Path:     path,
		RawPath:  path,
		RawQuery: q.Encode(),
	}
	return templ.SafeURL(u.String())
}

func Error(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	io.WriteString(w, msg)
}
