package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

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
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Content-Length", "16")

		// Act
		err := modifyResponse(r)
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
		}
		r.Header.Set("Content-Type", "text/html")
		r.Header.Set("Content-Length", "5")
		r.Header.Set("templ-skip-modify", "true")

		// Act
		err := modifyResponse(r)
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
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Length", "26")

		expectedString := insertScriptTagIntoBody(`<html><body></body></html>`)
		if !strings.Contains(expectedString, scriptTag) {
			t.Fatalf("expected the script tag to be inserted, but it wasn't: %q", expectedString)
		}

		// Act
		err := modifyResponse(r)
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
		}
		r.Header.Set("Content-Type", "application/json")
		// It's not actually gzipped here, but it doesn't matter, it shouldn't get that far.
		r.Header.Set("Content-Encoding", "gzip")
		// Similarly, this is not the actual length of the gzipped content.
		r.Header.Set("Content-Length", "16")

		// Act
		err := modifyResponse(r)
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

		expectedString := insertScriptTagIntoBody(body)

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
		}
		r.Header.Set("Content-Type", "text/html, charset=utf-8")
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Length", fmt.Sprintf("%d", expectedLength))

		// Act
		if err = modifyResponse(r); err != nil {
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

	t.Run("notify-proxy: sending POST request to /_templ/reload/events should receive reload sse event", func(t *testing.T) {
		// Arrange 1: create a test proxy server.
		dummyHandler := func(w http.ResponseWriter, r *http.Request) {}
		dummyServer := httptest.NewServer(http.HandlerFunc(dummyHandler))
		defer dummyServer.Close()

		u, err := url.Parse(dummyServer.URL)
		if err != nil {
			t.Fatalf("unexpected error parsing URL: %v", err)
		}
		handler := New("0.0.0.0", 0, u)
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
			// We expect the connection to be closed by the server: this is the only way to terminate the sse connection.
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
}
