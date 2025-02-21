package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ/cmd/templ/generatecmd/sse"
	"github.com/a-h/templ/internal/htmlfind"
	"github.com/andybalholm/brotli"
	"golang.org/x/net/html"

	_ "embed"
)

//go:embed script.js
var script string

type Handler struct {
	log    *slog.Logger
	URL    string
	Target *url.URL
	p      *httputil.ReverseProxy
	sse    *sse.Handler
}

func reloadScript(nonce string) *html.Node {
	script := &html.Node{
		Type: html.ElementNode,
		Data: "script",
		Attr: []html.Attribute{
			{Key: "src", Val: "/_templ/reload/script.js"},
		},
	}
	if nonce != "" {
		script.Attr = append(script.Attr, html.Attribute{Key: "nonce", Val: nonce})
	}
	return script
}

var ErrBodyNotFound = fmt.Errorf("body not found")

func insertScriptTagIntoBody(nonce, body string) (updated string, err error) {
	n, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return body, err
	}
	bodyNodes := htmlfind.All(n, htmlfind.Element("body"))
	if len(bodyNodes) == 0 {
		return body, ErrBodyNotFound
	}
	bodyNodes[0].AppendChild(reloadScript(nonce))
	buf := new(bytes.Buffer)
	if err = html.Render(buf, n); err != nil {
		return body, err
	}
	return buf.String(), nil
}

type passthroughWriteCloser struct {
	io.Writer
}

func (pwc passthroughWriteCloser) Close() error {
	return nil
}

const unsupportedContentEncoding = "Unsupported content encoding, hot reload script not inserted."

func (h *Handler) modifyResponse(r *http.Response) error {
	log := h.log.With(slog.String("url", r.Request.URL.String()))
	if r.Header.Get("templ-skip-modify") == "true" {
		log.Debug("Skipping response modification because templ-skip-modify header is set")
		return nil
	}
	if contentType := r.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		log.Debug("Skipping response modification because content type is not text/html", slog.String("content-type", contentType))
		return nil
	}

	// Set up readers and writers.
	newReader := func(in io.Reader) (out io.Reader, err error) {
		return in, nil
	}
	newWriter := func(out io.Writer) io.WriteCloser {
		return passthroughWriteCloser{out}
	}
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		newReader = func(in io.Reader) (out io.Reader, err error) {
			return gzip.NewReader(in)
		}
		newWriter = func(out io.Writer) io.WriteCloser {
			return gzip.NewWriter(out)
		}
	case "br":
		newReader = func(in io.Reader) (out io.Reader, err error) {
			return brotli.NewReader(in), nil
		}
		newWriter = func(out io.Writer) io.WriteCloser {
			return brotli.NewWriter(out)
		}
	case "":
		log.Debug("No content encoding header found")
	default:
		h.log.Warn(unsupportedContentEncoding, slog.String("encoding", r.Header.Get("Content-Encoding")))
	}

	// Read the encoded body.
	encr, err := newReader(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(encr)
	if err != nil {
		return err
	}

	// Update it.
	csp := r.Header.Get("Content-Security-Policy")
	updated, err := insertScriptTagIntoBody(parseNonce(csp), string(body))
	if err != nil {
		log.Warn("Unable to insert reload script", slog.Any("error", err))
		updated = string(body)
	}
	if len(updated) == len(body) {
		log.Debug("Reload script not inserted")
	} else {
		log.Debug("Reload script inserted")
	}

	// Encode the response.
	var buf bytes.Buffer
	encw := newWriter(&buf)
	_, err = encw.Write([]byte(updated))
	if err != nil {
		return err
	}
	err = encw.Close()
	if err != nil {
		return err
	}

	// Update the response.
	r.Body = io.NopCloser(&buf)
	r.ContentLength = int64(buf.Len())
	r.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	return nil
}

func parseNonce(csp string) (nonce string) {
outer:
	for _, rawDirective := range strings.Split(csp, ";") {
		parts := strings.Fields(rawDirective)
		if len(parts) < 2 {
			continue
		}
		if parts[0] != "script-src" {
			continue
		}
		for _, source := range parts[1:] {
			source = strings.TrimPrefix(source, "'")
			source = strings.TrimSuffix(source, "'")
			if strings.HasPrefix(source, "nonce-") {
				nonce = source[6:]
				break outer
			}
		}
	}
	return nonce
}

func New(log *slog.Logger, bind string, port int, target *url.URL) (h *Handler) {
	p := httputil.NewSingleHostReverseProxy(target)
	p.ErrorLog = stdlog.New(os.Stderr, "Proxy to target error: ", 0)
	p.Transport = &roundTripper{
		maxRetries:      20,
		initialDelay:    100 * time.Millisecond,
		backoffExponent: 1.5,
	}
	h = &Handler{
		log:    log,
		URL:    fmt.Sprintf("http://%s:%d", bind, port),
		Target: target,
		p:      p,
		sse:    sse.New(),
	}
	p.ModifyResponse = h.modifyResponse
	return h
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_templ/reload/script.js" {
		// Provides a script that reloads the page.
		w.Header().Add("Content-Type", "text/javascript")
		_, err := io.WriteString(w, script)
		if err != nil {
			fmt.Printf("failed to write script: %v\n", err)
		}
		return
	}
	if r.URL.Path == "/_templ/reload/events" {
		switch r.Method {
		case http.MethodGet:
			// Provides a list of messages including a reload message.
			p.sse.ServeHTTP(w, r)
			return
		case http.MethodPost:
			// Send a reload message to all connected clients.
			p.sse.Send("message", "reload")
			return
		}
		http.Error(w, "only GET or POST method allowed", http.StatusMethodNotAllowed)
		return
	}
	p.p.ServeHTTP(w, r)
}

func (p *Handler) SendSSE(eventType string, data string) {
	p.sse.Send(eventType, data)
}

type roundTripper struct {
	maxRetries      int
	initialDelay    time.Duration
	backoffExponent float64
}

func (rt *roundTripper) setShouldSkipResponseModificationHeader(r *http.Request, resp *http.Response) {
	// Instruct the modifyResponse function to skip modifying the response if the
	// HTTP request has come from HTMX.
	if r.Header.Get("HX-Request") != "true" {
		return
	}
	resp.Header.Set("templ-skip-modify", "true")
}

func (rt *roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Read and buffer the body.
	var bodyBytes []byte
	if r.Body != nil && r.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close()
	}

	// Retry logic.
	var resp *http.Response
	var err error
	for retries := 0; retries < rt.maxRetries; retries++ {
		// Clone the request and set the body.
		req := r.Clone(r.Context())
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Execute the request.
		resp, err = http.DefaultTransport.RoundTrip(req)
		if err != nil {
			time.Sleep(rt.initialDelay * time.Duration(math.Pow(rt.backoffExponent, float64(retries))))
			continue
		}

		rt.setShouldSkipResponseModificationHeader(r, resp)

		return resp, nil
	}

	return nil, fmt.Errorf("max retries reached: %q", r.URL.String())
}

func NotifyProxy(host string, port int) error {
	urlStr := fmt.Sprintf("http://%s:%d/_templ/reload/events", host, port)
	req, err := http.NewRequest(http.MethodPost, urlStr, nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}
