package templ_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestHandler(t *testing.T) {
	hello := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return nil
	})
	errorComponent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return errors.New("handler error")
	})

	tests := []struct {
		name             string
		input            *templ.ComponentHandler
		expectedStatus   int
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "handlers return OK by default",
			input:            templ.Handler(hello),
			expectedStatus:   http.StatusOK,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers return OK by default",
			input:            templ.Handler(templ.Raw(`♠ ‘ &spades; &#8216;`)),
			expectedStatus:   http.StatusOK,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "♠ ‘ &spades; &#8216;",
		},
		{
			name:             "handlers can be configured to return an alternative status code",
			input:            templ.Handler(hello, templ.WithStatus(http.StatusNotFound)),
			expectedStatus:   http.StatusNotFound,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers can be configured to return an alternative status code and content type",
			input:            templ.Handler(hello, templ.WithStatus(http.StatusOK), templ.WithContentType("text/csv")),
			expectedStatus:   http.StatusOK,
			expectedMIMEType: "text/csv",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers that fail return a 500 error",
			input:            templ.Handler(errorComponent),
			expectedStatus:   http.StatusInternalServerError,
			expectedMIMEType: "text/plain; charset=utf-8",
			expectedBody:     "templ: failed to render template\n",
		},
		{
			name: "error handling can be customised",
			input: templ.Handler(errorComponent, templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
				// Because the error is received, it's possible to log the detail of the request.
				// log.Printf("template render error for %v %v: %v", r.Method, r.URL.String(), err)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					if _, err := io.WriteString(w, "custom body"); err != nil {
						t.Fatalf("failed to write string: %v", err)
					}
				})
			})),
			expectedStatus:   http.StatusBadRequest,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "custom body",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			tt.input.ServeHTTP(w, r)
			if got := w.Result().StatusCode; tt.expectedStatus != got {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, got)
			}
			if mimeType := w.Result().Header.Get("Content-Type"); tt.expectedMIMEType != mimeType {
				t.Errorf("expected content-type %s, got %s", tt.expectedMIMEType, mimeType)
			}
			body, err := io.ReadAll(w.Result().Body)
			if err != nil {
				t.Errorf("failed to read body: %v", err)
			}
			if diff := cmp.Diff(tt.expectedBody, string(body)); diff != "" {
				t.Error(diff)
			}
		})
	}

	t.Run("streaming mode allows responses to be flushed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			// Write part 1.
			if _, err := io.WriteString(w, "Part 1"); err != nil {
				return err
			}
			// Flush.
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			// Check partial response.
			wr := w.(*httptest.ResponseRecorder)
			actualBody := wr.Body.String()
			if diff := cmp.Diff("Part 1", actualBody); diff != "" {
				t.Error(diff)
			}
			// Write part 2.
			if _, err := io.WriteString(w, "\nPart 2"); err != nil {
				return err
			}
			return nil
		})

		templ.Handler(component, templ.WithStatus(http.StatusCreated), templ.WithStreaming()).ServeHTTP(w, r)
		if got := w.Result().StatusCode; http.StatusCreated != got {
			t.Errorf("expected status %d, got %d", http.StatusCreated, got)
		}
		if mimeType := w.Result().Header.Get("Content-Type"); "text/html; charset=utf-8" != mimeType {
			t.Errorf("expected content-type %s, got %s", "text/html; charset=utf-8", mimeType)
		}
		body, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
		}
		if diff := cmp.Diff("Part 1\nPart 2", string(body)); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("streaming mode handles errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		expectedErr := errors.New("streaming error")

		component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			if _, err := io.WriteString(w, "Body"); err != nil {
				return err
			}
			return expectedErr
		})

		var errorHandlerCalled bool
		errorHandler := func(r *http.Request, err error) http.Handler {
			if expectedErr != err {
				t.Errorf("expected error %v, got %v", expectedErr, err)
			}
			errorHandlerCalled = true
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This will be ignored, because the header has already been written.
				w.WriteHeader(http.StatusBadRequest)
				// This will be written, but will be appended to the written body.
				if _, err := io.WriteString(w, "Error message"); err != nil {
					t.Errorf("failed to write error message: %v", err)
				}
			})
		}

		h := templ.Handler(component,
			templ.WithStatus(http.StatusCreated),
			templ.WithStreaming(),
			templ.WithErrorHandler(errorHandler),
		)
		h.ServeHTTP(w, r)

		if !errorHandlerCalled {
			t.Error("expected error handler to be called")
		}
		// Expect the status code to be 201, not 400, because in streaming mode,
		// we have to write the header before we can call the error handler.
		if actualResponseCode := w.Result().StatusCode; http.StatusCreated != actualResponseCode {
			t.Errorf("expected status %d, got %d", http.StatusCreated, actualResponseCode)
		}
		// Expect the body to be "BodyError message", not just "Error message" because
		// in streaming mode, we've already written part of the body to the response, unlike in
		// standard mode where the body is written to a buffer before the response is written,
		// ensuring that partial responses are not sent.
		actualBody, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
		}
		if diff := cmp.Diff("BodyError message", string(actualBody)); diff != "" {
			t.Error(diff)
		}
	})
}
