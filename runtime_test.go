package templ

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRenderedCSSClassesFromContext(t *testing.T) {
	ctx := context.Background()
	ctx, classes := RenderedCSSClassesFromContext(ctx)
	if classes.Contains("test") {
		t.Fatalf("before the classes have been set, test should not be set")
	}
	classes.Add("test")
	if !classes.Contains("test") {
		t.Errorf("expected 'test' to be present in the context, after setting")
	}
	_, updatedClasses := RenderedCSSClassesFromContext(ctx)
	if !updatedClasses.Contains("test") {
		t.Errorf("expected 'test' to be present in the context with new context, but it wasn't")
	}
	if classes != updatedClasses {
		t.Errorf("expected %v to be the same as %v", classes, updatedClasses)
	}
}

func TestCSSHandler(t *testing.T) {
	var tests = []struct {
		name             string
		input            []ComponentCSSClass
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "no classes",
			input:            nil,
			expectedMIMEType: "text/css",
			expectedBody:     "",
		},
		{
			name:             "classes are rendered",
			input:            []ComponentCSSClass{{ID: "className", Class: SafeCSS(".className{background-color:white;}")}},
			expectedMIMEType: "text/css",
			expectedBody:     ".className{background-color:white;}",
		},
		{
			name: "classes are rendered",
			input: []ComponentCSSClass{
				{ID: "classA", Class: SafeCSS(".classA{background-color:white;}")},
				{ID: "classB", Class: SafeCSS(".classB{background-color:green;}")},
			},
			expectedMIMEType: "text/css",
			expectedBody:     ".classA{background-color:white;}.classB{background-color:green;}",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := NewCSSHandler(tt.input...)
			h.ServeHTTP(w, &http.Request{})
			if diff := cmp.Diff(tt.expectedMIMEType, w.Header().Get("Content-Type")); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tt.expectedBody, w.Body.String()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestCSSMiddleware(t *testing.T) {
	pageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, classes := RenderedCSSClassesFromContext(r.Context())
		io.WriteString(w, "classes: "+strings.Join(classes.All(), " "))
	})
	c1 := ComponentCSSClass{
		ID:    "c1",
		Class: ".c1{color:red}",
	}
	c2 := ComponentCSSClass{
		ID:    "c2",
		Class: ".c2{color:blue}",
	}

	var tests = []struct {
		name             string
		input            *http.Request
		handler          http.Handler
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "accessing /style/templ.css renders CSS, even if it's empty",
			input:            httptest.NewRequest("GET", "/styles/templ.css", nil),
			handler:          NewCSSMiddleware(pageHandler),
			expectedMIMEType: "text/css",
			expectedBody:     "",
		},
		{
			name:             "accessing /style/templ.css renders CSS that includes the classes",
			input:            httptest.NewRequest("GET", "/styles/templ.css", nil),
			handler:          NewCSSMiddleware(pageHandler, c1, c2),
			expectedMIMEType: "text/css",
			expectedBody:     ".c1{color:red}.c2{color:blue}",
		},
		{
			name:             "the pageHandler can find out which CSS classes to skip rendering, even if there are none",
			input:            httptest.NewRequest("GET", "/index.html", nil),
			handler:          NewCSSMiddleware(pageHandler),
			expectedMIMEType: "text/plain; charset=utf-8",
			expectedBody:     "classes: ",
		},
		{
			name:             "the pageHandler can find out which CSS classes to skip rendering",
			input:            httptest.NewRequest("GET", "/index.html", nil),
			handler:          NewCSSMiddleware(pageHandler, c1, c2),
			expectedMIMEType: "text/plain; charset=utf-8",
			expectedBody:     "classes: c1 c2",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.handler.ServeHTTP(w, tt.input)
			if diff := cmp.Diff(tt.expectedMIMEType, w.Header().Get("Content-Type")); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tt.expectedBody, w.Body.String()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestRenderCSS(t *testing.T) {
	c1 := ComponentCSSClass{
		ID:    "c1",
		Class: ".c1{color:red}",
	}
	c2 := ComponentCSSClass{
		ID:    "c2",
		Class: ".c2{color:blue}",
	}

	var tests = []struct {
		name     string
		toIgnore []string
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name:     "if something outside the expected is ignored, if has no effect",
			toIgnore: []string{"c3"},
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []string{"c1"},
			expected: `<style type="text/css">.c2{color:blue}</style>`,
		},
		{
			name:     "if all are ignored, not even style tags are rendered",
			toIgnore: []string{"c1", "c2", "c3"},
			expected: ``,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, classes := RenderedCSSClassesFromContext(ctx)
			for _, toIgnore := range tt.toIgnore {
				classes.Add(toIgnore)
			}

			b := new(bytes.Buffer)
			err := RenderCSS(ctx, b, []CSSClass{c1, c2})
			if err != nil {
				t.Fatalf("failed to render CSS: %v", err)
			}

			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestClassSanitization(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{
			input:    `safe`,
			expected: `safe`,
		},
		{
			input:    `!unsafe`,
			expected: string(fallbackClassName),
		},
		{
			input:    `</style>`,
			expected: string(fallbackClassName),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			actual := Class(tt.input)
			if actual.ClassName() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual.ClassName())
			}
		})
	}

}

func TestHandler(t *testing.T) {
	hello := ComponentFunc(func(ctx context.Context, w io.Writer) error {
		io.WriteString(w, "Hello")
		return nil
	})
	errorComponent := ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return errors.New("handler error")
	})

	var tests = []struct {
		name           string
		input          *ComponentHandler
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "handlers return OK by default",
			input:          Handler(hello),
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello",
		},
		{
			name:           "handlers can be configured to return an alternative status code",
			input:          Handler(hello, WithStatus(http.StatusNotFound)),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Hello",
		},
		{
			name:           "handlers that fail return a 500 error",
			input:          Handler(errorComponent),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   componentHandlerErrorMessage + "\n",
		},
		{
			name: "error handling can be customised",
			input: Handler(errorComponent, WithErrorHandler(func(r *http.Request, err error) http.Handler {
				// Because the error is received, it's possible to log the detail of the request.
				// log.Printf("template render error for %v %v: %v", r.Method, r.URL.String(), err)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					io.WriteString(w, "custom body")
				})
			})),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "custom body",
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
			body, err := io.ReadAll(w.Result().Body)
			if err != nil {
				t.Errorf("failed to read body: %v", err)
			}
			if diff := cmp.Diff(tt.expectedBody, string(body)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
