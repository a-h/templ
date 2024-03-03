package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
}
