package main

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-h/templ"
)

func TestPageStreamsBeforeDataIsReady(t *testing.T) {
	data := make(chan string, 1)
	server := httptest.NewServer(templ.Handler(Page(data), templ.WithStreaming()))
	defer server.Close()
	defer close(data)

	client := server.Client()
	client.Timeout = 5 * time.Second
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("receive response before sending data: %v", err)
	}
	defer resp.Body.Close()

	const header = "<!doctype html><html><head><title>Page</title></head><body><h1>Page</h1>"
	readChunk(t, resp.Body, header)

	data <- "Part 1"
	readChunk(t, resp.Body, "<div>Part 1</div>")
}

func readChunk(t *testing.T, body io.Reader, want string) {
	t.Helper()
	got := make([]byte, len(want))
	if _, err := io.ReadFull(body, got); err != nil {
		t.Fatalf("receive %q: %v", want, err)
	}
	if string(got) != want {
		t.Fatalf("received %q, want %q", got, want)
	}
}
