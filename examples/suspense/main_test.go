package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type disconnectingResponseWriter struct {
	header       http.Header
	connectedFor time.Duration
	connectedAt  time.Time
	body         strings.Builder
}

func (w *disconnectingResponseWriter) Header() http.Header {
	return w.header
}

func (w *disconnectingResponseWriter) Write(p []byte) (n int, err error) {
	if w.connectedAt.IsZero() {
		w.connectedAt = time.Now()
	}
	if time.Since(w.connectedAt) >= w.connectedFor {
		return 0, errors.New("client disconnected")
	}
	return w.body.Write(p)
}

func (w *disconnectingResponseWriter) WriteHeader(int) {}

type panickingResponseWriter struct {
	header http.Header
}

func (w *panickingResponseWriter) Header() http.Header {
	return w.header
}

func (w *panickingResponseWriter) Write([]byte) (int, error) {
	panic("write failed")
}

func (w *panickingResponseWriter) WriteHeader(int) {}

// serve calls handleSuspense, and recovers from a panic, as net/http does.
func serve(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()
	handleSuspense(w, r)
}

func TestHandleSuspenseStreamsEverySlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		w := httptest.NewRecorder()
		handleSuspense(w, httptest.NewRequest(http.MethodGet, "/", nil))

		for _, expected := range []string{
			`<div slot="a"><div>Component A.</div></div>`,
			`<div slot="b"><div>Component B.</div></div>`,
			`<div slot="c"><div>Component C.</div></div>`,
		} {
			if !strings.Contains(w.Body.String(), expected) {
				t.Errorf("expected body to contain %q, got:\n%s", expected, w.Body.String())
			}
		}
	})
}

func TestHandleSuspenseReturnsBeforeTheSidebarDelayWhenTheRequestIsCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*1500)
		defer cancel()

		start := time.Now()
		handleSuspense(httptest.NewRecorder(), httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil))

		if elapsed := time.Since(start); elapsed >= time.Second*3 {
			t.Errorf("expected the handler to return before the 3s sidebar delay, took %v", elapsed)
		}
	})
}

func TestHandleSuspenseSlotGoroutinesExit(t *testing.T) {
	tests := []struct {
		name        string
		w           http.ResponseWriter
		cancelAfter time.Duration
	}{
		{
			name: "slot goroutines exit after the client reads every slot",
			w:    httptest.NewRecorder(),
		},
		{
			name: "slot goroutines exit when the client disconnects before any slot is read",
			w:    &disconnectingResponseWriter{header: http.Header{}},
		},
		{
			name: "slot goroutines exit when the client disconnects after some slots are read",
			w:    &disconnectingResponseWriter{header: http.Header{}, connectedFor: time.Millisecond * 1500},
		},
		{
			name:        "slot goroutines exit when the request is cancelled after some slots are read",
			w:           httptest.NewRecorder(),
			cancelAfter: time.Millisecond * 1500,
		},
		{
			name: "slot goroutines exit when rendering panics",
			w:    &panickingResponseWriter{header: http.Header{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// synctest fails the test if any goroutine started within the bubble
			// is still blocked when the function returns.
			synctest.Test(t, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				if tt.cancelAfter > 0 {
					ctx, cancel := context.WithTimeout(r.Context(), tt.cancelAfter)
					defer cancel()
					r = r.WithContext(ctx)
				}
				serve(tt.w, r)

				// Advance the fake clock past every slot delay.
				time.Sleep(time.Minute)
				synctest.Wait()
			})
		})
	}
}
