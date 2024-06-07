package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/google/go-cmp/cmp"
)

func TestRoundTripper(t *testing.T) {
	t.Run("if the HX-Request header is present, set the templ-skip-modify header on the response", func(t *testing.T) {
		rt := &roundTripper{}
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error creating request: %v", err)
		}
		req.Header.Set("HX-Request", "true")
		resp := &http.Response{Header: make(http.Header)}
		rt.setShouldSkipResponseModificationHeader(req, resp)
		if resp.Header.Get("templ-skip-modify") != "true" {
			t.Errorf("expected templ-skip-modify header to be true, got %v", resp.Header.Get("templ-skip-modify"))
		}
	})
	t.Run("if the HX-Request header is not present, do not set the templ-skip-modify header on the response", func(t *testing.T) {
		rt := &roundTripper{}
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error creating request: %v", err)
		}
		resp := &http.Response{Header: make(http.Header)}
		rt.setShouldSkipResponseModificationHeader(req, resp)
		if resp.Header.Get("templ-skip-modify") != "" {
			t.Errorf("expected templ-skip-modify header to be empty, got %v", resp.Header.Get("templ-skip-modify"))
		}
	})
}

func TestProxy(t *testing.T) {
	t.Run("plain: non-html content is not modified", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`{"key": "value"}`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Content-Length", "16")

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != "16" {
			t.Errorf("expected content length to be 16, got %v", r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(`{"key": "value"}`, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("plain: if the response contains templ-skip-modify header, it is not modified", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`Hello`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html")
		r.Header.Set("Content-Length", "5")
		r.Header.Set("templ-skip-modify", "true")

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != "5" {
			t.Errorf("expected content length to be 5, got %v", r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(`Hello`, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("plain: body tags get the script inserted", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`<html><body></body></html>`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Length", "26")

		expectedString := insertScriptTagIntoBody("", `<html><body></body></html>`)
		if !strings.Contains(expectedString, getScriptTag("")) {
			t.Fatalf("expected the script tag to be inserted, but it wasn't: %q", expectedString)
		}

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != fmt.Sprintf("%d", len(expectedString)) {
			t.Errorf("expected content length to be %d, got %v", len(expectedString), r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(expectedString, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("plain: body tags get the script inserted with nonce", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`<html><body></body></html>`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Length", "26")
		const nonce = "this-is-the-nonce"
		r.Header.Set("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-%s'", nonce))

		expectedString := insertScriptTagIntoBody(nonce, `<html><body></body></html>`)
		if !strings.Contains(expectedString, getScriptTag(nonce)) {
			t.Fatalf("expected the script tag to be inserted, but it wasn't: %q", expectedString)
		}

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != fmt.Sprintf("%d", len(expectedString)) {
			t.Errorf("expected content length to be %d, got %v", len(expectedString), r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(expectedString, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("plain: body tags get the script inserted ignoring js with body tags", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`<html><body><script>console.log("<body></body>")</script></body></html>`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Length", "26")

		expectedString := insertScriptTagIntoBody("", `<html><body><script>console.log("<body></body>")</script></body></html>`)
		if !strings.Contains(expectedString, getScriptTag("")) {
			t.Fatalf("expected the script tag to be inserted, but it wasn't: %q", expectedString)
		}
		if !strings.Contains(expectedString, `console.log("<body></body>")`) {
			t.Fatalf("expected the script tag to be inserted, but mangled the html: %q", expectedString)
		}

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != fmt.Sprintf("%d", len(expectedString)) {
			t.Errorf("expected content length to be %d, got %v", len(expectedString), r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(expectedString, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("gzip: non-html content is not modified", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(strings.NewReader(`{"key": "value"}`)),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "application/json")
		// It's not actually gzipped here, but it doesn't matter, it shouldn't get that far.
		r.Header.Set("Content-Encoding", "gzip")
		// Similarly, this is not the actual length of the gzipped content.
		r.Header.Set("Content-Length", "16")

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != "16" {
			t.Errorf("expected content length to be 16, got %v", r.Header.Get("Content-Length"))
		}
		actualBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(`{"key": "value"}`, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("gzip: body tags get the script inserted", func(t *testing.T) {
		// Arrange
		body := `<html><body></body></html>`
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write([]byte(body))
		if err != nil {
			t.Fatalf("unexpected error writing gzip: %v", err)
		}
		gzw.Close()

		expectedString := insertScriptTagIntoBody("", body)

		var expectedBytes bytes.Buffer
		gzw = gzip.NewWriter(&expectedBytes)
		_, err = gzw.Write([]byte(expectedString))
		if err != nil {
			t.Fatalf("unexpected error writing gzip: %v", err)
		}
		gzw.Close()
		expectedLength := len(expectedBytes.Bytes())

		r := &http.Response{
			Body:   io.NopCloser(&buf),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Length", fmt.Sprintf("%d", expectedLength))

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err = h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != fmt.Sprintf("%d", expectedLength) {
			t.Errorf("expected content length to be %d, got %v", expectedLength, r.Header.Get("Content-Length"))
		}

		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		actualBody, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(expectedString, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("brotli: body tags get the script inserted", func(t *testing.T) {
		// Arrange
		body := `<html><body></body></html>`
		var buf bytes.Buffer
		brw := brotli.NewWriter(&buf)
		_, err := brw.Write([]byte(body))
		if err != nil {
			t.Fatalf("unexpected error writing gzip: %v", err)
		}
		brw.Close()

		expectedString := insertScriptTagIntoBody("", body)

		var expectedBytes bytes.Buffer
		brw = brotli.NewWriter(&expectedBytes)
		_, err = brw.Write([]byte(expectedString))
		if err != nil {
			t.Fatalf("unexpected error writing gzip: %v", err)
		}
		brw.Close()
		expectedLength := len(expectedBytes.Bytes())

		r := &http.Response{
			Body:   io.NopCloser(&buf),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Encoding", "br")
		r.Header.Set("Content-Length", fmt.Sprintf("%d", expectedLength))

		// Act
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err = h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if r.Header.Get("Content-Length") != fmt.Sprintf("%d", expectedLength) {
			t.Errorf("expected content length to be %d, got %v", expectedLength, r.Header.Get("Content-Length"))
		}

		actualBody, err := io.ReadAll(brotli.NewReader(r.Body))
		if err != nil {
			t.Fatalf("unexpected error reading response: %v", err)
		}
		if diff := cmp.Diff(expectedString, string(actualBody)); diff != "" {
			t.Errorf("unexpected response body (-got +want):\n%s", diff)
		}
	})
	t.Run("notify-proxy: sending POST request to /_templ/reload/events should receive reload sse event", func(t *testing.T) {
		// Arrange 1: create a test proxy server.
		dummyHandler := func(w http.ResponseWriter, r *http.Request) {}
		dummyServer := httptest.NewServer(http.HandlerFunc(dummyHandler))
		defer dummyServer.Close()

		u, err := url.Parse(dummyServer.URL)
		if err != nil {
			t.Fatalf("unexpected error parsing URL: %v", err)
		}
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		handler := New(log, "0.0.0.0", 0, u)
		proxyServer := httptest.NewServer(handler)
		defer proxyServer.Close()

		u2, err := url.Parse(proxyServer.URL)
		if err != nil {
			t.Fatalf("unexpected error parsing URL: %v", err)
		}
		port, err := strconv.Atoi(u2.Port())
		if err != nil {
			t.Fatalf("unexpected error parsing port: %v", err)
		}

		// Arrange 2: start a goroutine to listen for sse events.
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		errChan := make(chan error)
		sseRespCh := make(chan string)
		sseListening := make(chan bool) // Coordination channel that ensures the SSE listener is started before notifying the proxy.
		go func() {
			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/_templ/reload/events", proxyServer.URL), nil)
			if err != nil {
				errChan <- err
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()

			sseListening <- true
			lines := []string{}
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				if scanner.Text() == "data: reload" {
					sseRespCh <- strings.Join(lines, "\n")
					return
				}
			}
			err = scanner.Err()
			if err != nil {
				errChan <- err
				return
			}
		}()

		// Act: notify the proxy.
		select { // Either SSE is listening or an error occurred.
		case <-sseListening:
			err = NotifyProxy(u2.Hostname(), port)
			if err != nil {
				t.Fatalf("unexpected error notifying proxy: %v", err)
			}
		case err := <-errChan:
			if err == nil {
				t.Fatalf("unexpected sse response: %v", err)
			}
		}

		// Assert.
		select { // Either SSE has a expected response or an error or timeout occurred.
		case resp := <-sseRespCh:
			if !strings.Contains(resp, "event: message\ndata: reload") {
				t.Errorf("expected sse reload event to be received, got: %q", resp)
			}
		case err := <-errChan:
			if err == nil {
				t.Fatalf("unexpected sse response: %v", err)
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for sse response")
		}
	})
	t.Run("unsupported encodings result in a warning", func(t *testing.T) {
		// Arrange
		r := &http.Response{
			Body:   io.NopCloser(bytes.NewReader([]byte("<p>Data</p>"))),
			Header: make(http.Header),
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
				},
			},
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Encoding", "weird-encoding")

		// Act
		lh := newTestLogHandler(slog.LevelInfo)
		log := slog.New(lh)
		h := New(log, "127.0.0.1", 7474, &url.URL{Scheme: "http", Host: "example.com"})
		err := h.modifyResponse(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if len(lh.records) != 1 {
			var sb strings.Builder
			for _, record := range lh.records {
				sb.WriteString(record.Message)
				sb.WriteString("\n")
			}
			t.Fatalf("expected 1 log entry, but got %d: \n%s", len(lh.records), sb.String())
		}
		record := lh.records[0]
		if record.Message != unsupportedContentEncoding {
			t.Errorf("expected warning message %q, got %q", unsupportedContentEncoding, record.Message)
		}
		if record.Level != slog.LevelWarn {
			t.Errorf("expected warning, got level %v", record.Level)
		}
	})
}

func newTestLogHandler(level slog.Level) *testLogHandler {
	return &testLogHandler{
		m:       new(sync.Mutex),
		records: nil,
		level:   level,
	}
}

type testLogHandler struct {
	m       *sync.Mutex
	records []slog.Record
	level   slog.Level
}

func (h *testLogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *testLogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.m.Lock()
	defer h.m.Unlock()
	if r.Level < h.level {
		return nil
	}
	h.records = append(h.records, r)
	return nil
}

func (h *testLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestParseNonce(t *testing.T) {
	for _, tc := range []struct {
		name     string
		csp      string
		expected string
	}{
		{
			name:     "empty csp",
			csp:      "",
			expected: "",
		},
		{
			name:     "simple csp",
			csp:      "script-src 'nonce-oLhVst3hTAcxI734qtB0J9Qc7W4qy09C'",
			expected: "oLhVst3hTAcxI734qtB0J9Qc7W4qy09C",
		},
		{
			name:     "simple csp without single quote",
			csp:      "script-src nonce-oLhVst3hTAcxI734qtB0J9Qc7W4qy09C",
			expected: "oLhVst3hTAcxI734qtB0J9Qc7W4qy09C",
		},
		{
			name:     "complete csp",
			csp:      "default-src 'self'; frame-ancestors 'self'; form-action 'self'; script-src 'strict-dynamic' 'nonce-4VOtk0Uo1l7pwtC';",
			expected: "4VOtk0Uo1l7pwtC",
		},
		{
			name:     "mdn example 1",
			csp:      "default-src 'self'",
			expected: "",
		},
		{
			name:     "mdn example 2",
			csp:      "default-src 'self' *.trusted.com",
			expected: "",
		},
		{
			name:     "mdn example 3",
			csp:      "default-src 'self'; img-src *; media-src media1.com media2.com; script-src userscripts.example.com",
			expected: "",
		},
		{
			name:     "mdn example 3 multiple sources",
			csp:      "default-src 'self'; img-src *; media-src media1.com media2.com; script-src userscripts.example.com foo.com 'strict-dynamic' 'nonce-4VOtk0Uo1l7pwtC'",
			expected: "4VOtk0Uo1l7pwtC",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nonce := parseNonce(tc.csp)
			if nonce != tc.expected {
				t.Errorf("expected nonce to be %s, but got %s", tc.expected, nonce)
			}
		})
	}
}
