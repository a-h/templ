package templ_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestCSSHandler(t *testing.T) {
	tests := []struct {
		name             string
		input            []templ.CSSClass
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
			input:            []templ.CSSClass{templ.ComponentCSSClass{ID: "className", Class: templ.SafeCSS(".className{background-color:white;}")}},
			expectedMIMEType: "text/css",
			expectedBody:     ".className{background-color:white;}",
		},
		{
			name: "classes are rendered",
			input: []templ.CSSClass{
				templ.ComponentCSSClass{ID: "classA", Class: templ.SafeCSS(".classA{background-color:white;}")},
				templ.ComponentCSSClass{ID: "classB", Class: templ.SafeCSS(".classB{background-color:green;}")},
			},
			expectedMIMEType: "text/css",
			expectedBody:     ".classA{background-color:white;}.classB{background-color:green;}",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := templ.NewCSSHandler(tt.input...)
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
		if _, err := io.WriteString(w, "Hello, World!"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
	})
	c1 := templ.ComponentCSSClass{
		ID:    "c1",
		Class: ".c1{color:red}",
	}
	c2 := templ.ComponentCSSClass{
		ID:    "c2",
		Class: ".c2{color:blue}",
	}

	tests := []struct {
		name             string
		input            *http.Request
		handler          http.Handler
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "accessing /style/templ.css renders CSS, even if it's empty",
			input:            httptest.NewRequest("GET", "/styles/templ.css", nil),
			handler:          templ.NewCSSMiddleware(pageHandler),
			expectedMIMEType: "text/css",
			expectedBody:     "",
		},
		{
			name:             "accessing /style/templ.css renders CSS that includes the classes",
			input:            httptest.NewRequest("GET", "/styles/templ.css", nil),
			handler:          templ.NewCSSMiddleware(pageHandler, c1, c2),
			expectedMIMEType: "text/css",
			expectedBody:     ".c1{color:red}.c2{color:blue}",
		},
		{
			name:             "the pageHandler is rendered",
			input:            httptest.NewRequest("GET", "/index.html", nil),
			handler:          templ.NewCSSMiddleware(pageHandler, c1, c2),
			expectedMIMEType: "text/plain; charset=utf-8",
			expectedBody:     "Hello, World!",
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
	c1 := templ.ComponentCSSClass{
		ID:    "c1",
		Class: ".c1{color:red}",
	}
	c2 := templ.ComponentCSSClass{
		ID:    "c2",
		Class: ".c2{color:blue}",
	}

	tests := []struct {
		name     string
		toIgnore []any
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name: "if something outside the expected is ignored, if has no effect",
			toIgnore: []any{
				templ.ComponentCSSClass{
					ID:    "c3",
					Class: templ.SafeCSS(".c3{color:yellow}"),
				},
			},
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []any{c1},
			expected: `<style type="text/css">.c2{color:blue}</style>`,
		},
		{
			name: "if all are ignored, not even style tags are rendered",
			toIgnore: []any{
				c1,
				c2,
				templ.ComponentCSSClass{
					ID:    "c3",
					Class: templ.SafeCSS(".c3{color:yellow}"),
				},
			},
			expected: ``,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := new(bytes.Buffer)

			// Render twice, reusing the same context so that there's a memory of which classes have been rendered.
			ctx = templ.InitializeContext(ctx)
			err := templ.RenderCSSItems(ctx, b, tt.toIgnore...)
			if err != nil {
				t.Fatalf("failed to render initial CSS: %v", err)
			}

			// Now render again to check that only the expected classes were rendered.
			b.Reset()
			err = templ.RenderCSSItems(ctx, b, []any{c1, c2}...)
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
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `safe`,
			expected: `safe`,
		},
		{
			input:    `safe-name`,
			expected: "safe-name",
		},
		{
			input:    `safe_name`,
			expected: "safe_name",
		},
		{
			input:    `!unsafe`,
			expected: "--templ-css-class-safe-name",
		},
		{
			input:    `</style>`,
			expected: "--templ-css-class-safe-name",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			actual := templ.Class(tt.input)
			if actual.ClassName() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual.ClassName())
			}
		})
	}
}

func TestClassesFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{
			name:     "safe constants are allowed",
			input:    []any{"a", "b", "c"},
			expected: "a b c",
		},
		{
			name:     "unsafe constants are filtered",
			input:    []any{"</style>", "b", "</style>"},
			expected: "--templ-css-class-safe-name b",
		},
		{
			name:     "legacy CSS types are supported",
			input:    []any{"a", templ.SafeClass("b"), templ.Class("c")},
			expected: "a b c",
		},
		{
			name: "CSS components are included in the output",
			input: []any{
				templ.ComponentCSSClass{ID: "classA", Class: templ.SafeCSS(".classA{background-color:white;}")},
				templ.ComponentCSSClass{ID: "classB", Class: templ.SafeCSS(".classB{background-color:green;}")},
				"c",
			},
			expected: "classA classB c",
		},
		{
			name: "optional classes can be applied with expressions",
			input: []any{
				"a",
				templ.ComponentCSSClass{ID: "classA", Class: templ.SafeCSS(".classA{background-color:white;}")},
				templ.ComponentCSSClass{ID: "classB", Class: templ.SafeCSS(".classB{background-color:green;}")},
				"c",
				map[string]bool{
					"a":      false,
					"classA": false,
					"classB": false,
					"c":      true,
					"d":      false,
				},
			},
			expected: "c",
		},
		{
			name: "unknown types for classes get rendered as --templ-css-class-unknown-type",
			input: []any{
				123,
				map[string]string{"test": "no"},
				false,
				"c",
			},
			expected: "--templ-css-class-unknown-type c",
		},
		{
			name: "string arrays are supported",
			input: []any{
				[]string{"a", "b", "c"},
				"d",
			},
			expected: "a b c d",
		},
		{
			name: "string arrays are checked for unsafe class names",
			input: []any{
				[]string{"a", "b", "c </style>"},
				"d",
			},
			expected: "a b c --templ-css-class-safe-name d",
		},
		{
			name: "strings are broken up",
			input: []any{
				"a </style>",
			},
			expected: "a --templ-css-class-safe-name",
		},
		{
			name: "if a templ.CSSClasses is passed in, the nested CSSClasses are extracted",
			input: []any{
				templ.Classes(
					"a",
					templ.SafeClass("b"),
					templ.Class("c"),
					templ.ComponentCSSClass{
						ID:    "d",
						Class: "{}",
					},
				),
			},
			expected: "a b c d",
		},
		{
			name: "kv types can be used to show or hide classes",
			input: []any{
				"a",
				templ.KV("b", true),
				"c",
				templ.KV("c", false),
			},
			expected: "a b",
		},
		{
			name: "an array of KV types can be used to show or hide classes",
			input: []any{
				"a",
				"c",
				[]templ.KeyValue[string, bool]{
					templ.KV("b", true),
					templ.KV("c", false),
					{"d", true},
				},
			},
			expected: "a b d",
		},
		{
			name: "the brackets on component CSS function calls can be elided",
			input: []any{
				func() templ.CSSClass {
					return templ.ComponentCSSClass{
						ID:    "a",
						Class: "",
					}
				},
			},
			expected: "a",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := templ.Classes(test.input...).String()
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	hello := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return nil
	})
	errorComponent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return errors.New("handler error")
	})

	tests := []struct {
		name           string
		input          *templ.ComponentHandler
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "handlers return OK by default",
			input:          templ.Handler(hello),
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello",
		},
		{
			name:           "handlers can be configured to return an alternative status code",
			input:          templ.Handler(hello, templ.WithStatus(http.StatusNotFound)),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Hello",
		},
		{
			name:           "handlers that fail return a 500 error",
			input:          templ.Handler(errorComponent),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "templ: failed to render template\n",
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
