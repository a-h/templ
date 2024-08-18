package templ_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
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

var cssInputs = []any{
	[]string{"a", "b"},          // []string
	"c",                         // string
	templ.ConstantCSSClass("d"), // ConstantCSSClass
	templ.ComponentCSSClass{ID: "e", Class: ".e{color:red}"}, // ComponentCSSClass
	map[string]bool{"f": true, "ff": false},                  // map[string]bool
	templ.KV[string, bool]("g", true),                        // KeyValue[string, bool]
	templ.KV[string, bool]("gg", false),                      // KeyValue[string, bool]
	[]templ.KeyValue[string, bool]{
		templ.KV("h", true),
		templ.KV("hh", false),
	}, // []KeyValue[string, bool]
	templ.KV[templ.CSSClass, bool](templ.ConstantCSSClass("i"), true),   // KeyValue[CSSClass, bool]
	templ.KV[templ.CSSClass, bool](templ.ConstantCSSClass("ii"), false), // KeyValue[CSSClass, bool]
	templ.KV[templ.ComponentCSSClass, bool](templ.ComponentCSSClass{
		ID:    "j",
		Class: ".j{color:red}",
	}, true), // KeyValue[ComponentCSSClass, bool]
	templ.KV[templ.ComponentCSSClass, bool](templ.ComponentCSSClass{
		ID:    "jj",
		Class: ".jj{color:red}",
	}, false), // KeyValue[ComponentCSSClass, bool]
	templ.CSSClasses{templ.ConstantCSSClass("k")},                             // CSSClasses
	func() templ.CSSClass { return templ.ConstantCSSClass("l") },              // func() CSSClass
	templ.CSSClass(templ.ConstantCSSClass("m")),                               // CSSClass
	customClass{name: "n"},                                                    // CSSClass
	[]templ.CSSClass{customClass{name: "n"}},                                  // []CSSClass
	templ.KV[templ.ConstantCSSClass, bool](templ.ConstantCSSClass("o"), true), // KeyValue[ConstantCSSClass, bool]
	[]templ.KeyValue[templ.ConstantCSSClass, bool]{
		templ.KV(templ.ConstantCSSClass("p"), true),
		templ.KV(templ.ConstantCSSClass("pp"), false),
	}, // []KeyValue[ConstantCSSClass, bool]
}

type customClass struct {
	name string
}

func (cc customClass) ClassName() string {
	return cc.name
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
		toRender []any
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			toRender: []any{c1, c2},
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
			toRender: []any{c1, c2},
			expected: `<style type="text/css">.c1{color:red}.c2{color:blue}</style>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []any{c1},
			toRender: []any{c1, c2},
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
			toRender: []any{c1, c2},
			expected: ``,
		},
		{
			name:     "CSS classes are rendered",
			toIgnore: nil,
			toRender: cssInputs,
			expected: `<style type="text/css">.e{color:red}.j{color:red}</style>`,
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
			err = templ.RenderCSSItems(ctx, b, tt.toRender...)
			if err != nil {
				t.Fatalf("failed to render CSS: %v", err)
			}

			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
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
			name:     "constants are allowed",
			input:    []any{"a", "b", "c", "</style>"},
			expected: "a b c </style>",
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
				[]string{"a", "b", "c", "</style>"},
				"d",
			},
			expected: "a b c </style> d",
		},
		{
			name: "strings are broken up",
			input: []any{
				"a </style>",
			},
			expected: "a </style>",
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
				templ.KV(templ.SafeClass("d"), true),
				templ.KV(templ.SafeClass("e"), false),
			},
			expected: "a b d",
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

type baseError struct {
	Value int
}

func (baseError) Error() string { return "base error" }

type nonMatchedError struct{}

func (nonMatchedError) Error() string { return "non matched error" }

func TestErrorWrapping(t *testing.T) {
	baseErr := baseError{
		Value: 1,
	}
	wrappedErr := templ.Error{Err: baseErr, Line: 1, Col: 2}
	t.Run("errors.Is() returns true for the base error", func(t *testing.T) {
		if !errors.Is(wrappedErr, baseErr) {
			t.Error("errors.Is() returned false for the base error")
		}
	})
	t.Run("errors.Is() returns false for a different error", func(t *testing.T) {
		if errors.Is(wrappedErr, errors.New("different error")) {
			t.Error("errors.Is() returned true for a different error")
		}
	})
	t.Run("errors.As() returns true for the base error", func(t *testing.T) {
		var err baseError
		if !errors.As(wrappedErr, &err) {
			t.Error("errors.As() returned false for the base error")
		}
		if err.Value != 1 {
			t.Errorf("errors.As() returned a different value: %v", err.Value)
		}
	})
	t.Run("errors.As() returns false for a different error", func(t *testing.T) {
		var err nonMatchedError
		if errors.As(wrappedErr, &err) {
			t.Error("errors.As() returned true for a different error")
		}
	})
}

func TestRawComponent(t *testing.T) {
	tests := []struct {
		name        string
		input       templ.Component
		expected    string
		expectedErr error
	}{
		{
			name:     "Raw content is not escaped",
			input:    templ.Raw("<div>Test &</div>"),
			expected: `<div>Test &</div>`,
		},
		{
			name:        "Raw will return errors first",
			input:       templ.Raw("", nil, errors.New("test error")),
			expected:    `<div>Test &</div>`,
			expectedErr: errors.New("test error"),
		},
		{
			name:     "Strings marked as safe are rendered without escaping",
			input:    templ.Raw(template.HTML("<div>")),
			expected: `<div>`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := tt.input.Render(context.Background(), b)
			if tt.expectedErr != nil {
				expected := tt.expectedErr.Error()
				actual := fmt.Sprintf("%v", err)
				if actual != expected {
					t.Errorf("expected error %q, got %q", expected, actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to render content: %v", err)
			}
			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
	t.Run("Raw does not require allocations", func(t *testing.T) {
		actualAllocs := testing.AllocsPerRun(4, func() {
			c := templ.Raw("<div>")
			if c == nil {
				t.Fatalf("unexpected nil value")
			}
		})
		if actualAllocs > 0 {
			t.Errorf("expected no allocs, got %v", actualAllocs)
		}
	})
}

var goTemplate = template.Must(template.New("example").Parse("<div>{{ . }}</div>"))

func TestGoHTMLComponents(t *testing.T) {
	t.Run("Go templates can be rendered as templ components", func(t *testing.T) {
		b := new(bytes.Buffer)
		err := templ.FromGoHTML(goTemplate, "Test &").Render(context.Background(), b)
		if err != nil {
			t.Fatalf("failed to render content: %v", err)
		}
		if diff := cmp.Diff("<div>Test &amp;</div>", b.String()); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("templ components can be rendered in Go templates", func(t *testing.T) {
		b := new(bytes.Buffer)
		c := templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
			_, err = io.WriteString(w, "<div>Unsanitized &</div>")
			return err
		})
		h, err := templ.ToGoHTML(context.Background(), c)
		if err != nil {
			t.Fatalf("failed to convert to Go HTML: %v", err)
		}
		if err = goTemplate.Execute(b, h); err != nil {
			t.Fatalf("failed to render content: %v", err)
		}
		if diff := cmp.Diff("<div><div>Unsanitized &</div></div>", b.String()); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("errors in ToGoHTML are returned", func(t *testing.T) {
		expectedErr := errors.New("test error")
		c := templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
			return expectedErr
		})
		_, err := templ.ToGoHTML(context.Background(), c)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err != expectedErr {
			t.Fatalf("expected error %q, got %q", expectedErr, err)
		}
	})
	t.Run("FromGoHTML does not require allocations", func(t *testing.T) {
		actualAllocs := testing.AllocsPerRun(4, func() {
			c := templ.FromGoHTML(goTemplate, "test &")
			if c == nil {
				t.Fatalf("unexpected nil value")
			}
		})
		if actualAllocs > 0 {
			t.Errorf("expected no allocs, got %v", actualAllocs)
		}
	})
	t.Run("ToGoHTML requires one allocation", func(t *testing.T) {
		expected := "<div>Unsanitized &</div>"
		c := templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
			_, err = io.WriteString(w, expected)
			return err
		})
		actualAllocs := testing.AllocsPerRun(4, func() {
			h, err := templ.ToGoHTML(context.Background(), c)
			if err != nil {
				t.Fatalf("failed to convert to Go HTML: %v", err)
			}
			if h != template.HTML(expected) {
				t.Fatalf("unexpected value")
			}
		})
		if actualAllocs > 1 {
			t.Errorf("expected 1 alloc, got %v", actualAllocs)
		}
	})
}

func TestNonce(t *testing.T) {
	ctx := context.Background()
	t.Run("returns empty string if not set", func(t *testing.T) {
		actual := templ.GetNonce(ctx)
		if actual != "" {
			t.Errorf("expected empty string got %q", actual)
		}
	})
	t.Run("returns value if one has been set", func(t *testing.T) {
		expected := "abc123"
		ctx := templ.WithNonce(context.Background(), expected)
		actual := templ.GetNonce(ctx)
		if actual != expected {
			t.Errorf("expected %q got %q", expected, actual)
		}
	})
}
