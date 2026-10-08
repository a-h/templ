package runtime

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/a-h/templ"
)

func TestRenderAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes templ.Attributes
		expected   string
	}{
		{
			name: "string attributes are rendered",
			attributes: templ.Attributes{
				"class": "test-class",
				"id":    "test-id",
			},
			expected: ` class="test-class" id="test-id"`,
		},
		{
			name: "integer types are rendered as strings",
			attributes: templ.Attributes{
				"int":   42,
				"int8":  int8(8),
				"int16": int16(16),
				"int32": int32(32),
				"int64": int64(64),
			},
			expected: ` int="42" int16="16" int32="32" int64="64" int8="8"`,
		},
		{
			name: "unsigned integer types are rendered as strings",
			attributes: templ.Attributes{
				"uint":    uint(42),
				"uint8":   uint8(8),
				"uint16":  uint16(16),
				"uint32":  uint32(32),
				"uint64":  uint64(64),
				"uintptr": uintptr(100),
			},
			expected: ` uint="42" uint16="16" uint32="32" uint64="64" uint8="8" uintptr="100"`,
		},
		{
			name: "float types are rendered as strings",
			attributes: templ.Attributes{
				"float32": float32(3.14),
				"float64": float64(2.718),
			},
			expected: ` float32="3.14" float64="2.718"`,
		},
		{
			name: "complex types are rendered as strings",
			attributes: templ.Attributes{
				"complex64":  complex64(1 + 2i),
				"complex128": complex128(3 + 4i),
			},
			expected: ` complex128="(3+4i)" complex64="(1+2i)"`,
		},
		{
			name: "boolean attributes are rendered correctly",
			attributes: templ.Attributes{
				"checked":  true,
				"disabled": false,
			},
			expected: ` checked`,
		},
		{
			name: "mixed types are rendered correctly",
			attributes: templ.Attributes{
				"class":  "button",
				"value":  42,
				"width":  float64(100.5),
				"hidden": false,
				"active": true,
			},
			expected: ` active class="button" value="42" width="100.5"`,
		},
		{
			name: "nil pointer attributes are not rendered",
			attributes: templ.Attributes{
				"optional": (*string)(nil),
				"visible":  (*bool)(nil),
			},
			expected: ``,
		},
		{
			name: "non-nil pointer attributes are rendered",
			attributes: templ.Attributes{
				"title":   new("test title"),
				"enabled": new(true),
			},
			expected: ` enabled title="test title"`,
		},
		{
			name: "numeric pointer types are rendered as strings",
			attributes: templ.Attributes{
				"int-ptr":        new(42),
				"int8-ptr":       new(int8(8)),
				"int16-ptr":      new(int16(16)),
				"int32-ptr":      new(int32(32)),
				"int64-ptr":      new(int64(64)),
				"uint-ptr":       new(uint(42)),
				"uint8-ptr":      new(uint8(8)),
				"uint16-ptr":     new(uint16(16)),
				"uint32-ptr":     new(uint32(32)),
				"uint64-ptr":     new(uint64(64)),
				"uintptr-ptr":    new(uintptr(100)),
				"float32-ptr":    new(float32(3.14)),
				"float64-ptr":    new(float64(2.718)),
				"complex64-ptr":  new(complex64(1 + 2i)),
				"complex128-ptr": new(complex128(3 + 4i)),
			},
			expected: ` complex128-ptr="(3+4i)" complex64-ptr="(1+2i)" float32-ptr="3.14" float64-ptr="2.718" int-ptr="42" int16-ptr="16" int32-ptr="32" int64-ptr="64" int8-ptr="8" uint-ptr="42" uint16-ptr="16" uint32-ptr="32" uint64-ptr="64" uint8-ptr="8" uintptr-ptr="100"`,
		},
		{
			name: "nil numeric pointer attributes are not rendered",
			attributes: templ.Attributes{
				"int-ptr":       (*int)(nil),
				"float32-ptr":   (*float32)(nil),
				"complex64-ptr": (*complex64)(nil),
			},
			expected: ``,
		},
		{
			name: "KeyValue[string, bool] attributes are rendered correctly",
			attributes: templ.Attributes{
				"data-value":  templ.KV("test-string", true),
				"data-hidden": templ.KV("ignored", false),
			},
			expected: ` data-value="test-string"`,
		},
		{
			name: "KeyValue[bool, bool] attributes are rendered correctly",
			attributes: templ.Attributes{
				"checked":  templ.KV(true, true),
				"disabled": templ.KV(false, true),
				"hidden":   templ.KV(true, false),
			},
			expected: ` checked`,
		},
		{
			name: "function bool attributes are rendered correctly",
			attributes: templ.Attributes{
				"enabled": func() bool { return true },
				"hidden":  func() bool { return false },
			},
			expected: ` enabled`,
		},
		{
			name: "mixed KeyValue and function attributes",
			attributes: templ.Attributes{
				"data-name": templ.KV("value", true),
				"active":    templ.KV(true, true),
				"dynamic":   func() bool { return true },
				"ignored":   templ.KV("ignored", false),
			},
			expected: ` active data-name="value" dynamic`,
		},
		{
			name: "string URL attributes are sanitized",
			attributes: templ.Attributes{
				"href":       "javascript:alert(1)",
				"formaction": "javascript:alert(2)",
				"xlink:href": "javascript:alert(3)",
			},
			expected: ` formaction="about:invalid#TemplFailedSanitizationURL" href="about:invalid#TemplFailedSanitizationURL" xlink:href="about:invalid#TemplFailedSanitizationURL"`,
		},
		{
			name: "URL attribute names are sanitized regardless of case",
			attributes: templ.Attributes{
				"HREF": "javascript:alert(1)",
			},
			expected: ` HREF="about:invalid#TemplFailedSanitizationURL"`,
		},
		{
			name: "URL attributes from KeyValue[string, bool] values are sanitized",
			attributes: templ.Attributes{
				"href": templ.KV("javascript:alert(1)", true),
			},
			expected: ` href="about:invalid#TemplFailedSanitizationURL"`,
		},
		{
			name: "string pointer URL attributes are sanitized",
			attributes: templ.Attributes{
				"href": new("javascript:alert(1)"),
			},
			expected: ` href="about:invalid#TemplFailedSanitizationURL"`,
		},
		{
			name: "safe URL strings are rendered unchanged",
			attributes: templ.Attributes{
				"href": "https://example.com/?a=1&b=2",
			},
			expected: ` href="https://example.com/?a=1&amp;b=2"`,
		},
		{
			name: "SafeURL values bypass URL sanitization",
			attributes: templ.Attributes{
				"href": templ.SafeURL("javascript:alert(1)"),
			},
			expected: ` href="javascript:alert(1)"`,
		},
		{
			name: "string event handlers are replaced",
			attributes: templ.Attributes{
				"onclick":          "alert(1)",
				"OnMouseOver":      "alert(2)",
				"hx-on:click":      "alert(3)",
				"hx-on-click":      "alert(4)",
				"data-hx-on:click": "alert(5)",
			},
			expected: ` OnMouseOver="` + templ.FailedSanitizationJS + `" data-hx-on:click="` + templ.FailedSanitizationJS + `" hx-on-click="` + templ.FailedSanitizationJS + `" hx-on:click="` + templ.FailedSanitizationJS + `" onclick="` + templ.FailedSanitizationJS + `"`,
		},
		{
			name: "event handlers from KeyValue[string, bool] values are replaced",
			attributes: templ.Attributes{
				"onclick": templ.KV("alert(1)", true),
			},
			expected: ` onclick="` + templ.FailedSanitizationJS + `"`,
		},
		{
			name: "ComponentScript event handlers are rendered",
			attributes: templ.Attributes{
				"onclick": templ.JSFuncCall("alert", "hello"),
			},
			expected: ` onclick="alert(&#34;hello&#34;)"`,
		},
		{
			name: "JSUnsafeFuncCall event handlers are rendered",
			attributes: templ.Attributes{
				"onclick": templ.JSUnsafeFuncCall("alert(1)"),
			},
			expected: ` onclick="alert(1)"`,
		},
		{
			name: "invalid attribute names are replaced",
			attributes: templ.Attributes{
				`x onmouseover=alert(1) y`: "value",
				`x"`:                       "value",
				`x>`:                       "value",
				`x/`:                       "value",
				"":                         "value",
			},
			expected: ` data-templ-failed-sanitization="value" data-templ-failed-sanitization="value" data-templ-failed-sanitization="value" data-templ-failed-sanitization="value" data-templ-failed-sanitization="value"`,
		},
		{
			name: "invalid boolean attribute names are replaced",
			attributes: templ.Attributes{
				`x onmouseover=alert(1) y`: true,
			},
			expected: ` data-templ-failed-sanitization`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := RenderAttributes(context.Background(), &buf, "div", tt.attributes)
			if err != nil {
				t.Fatalf("RenderAttributes failed: %v", err)
			}

			actual := buf.String()
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestRenderAttributesSrcdoc covers GHSA-94c2-xg24-pwhv: the browser parses
// the value of an iframe srcdoc attribute as an HTML document in the context
// of the page, so a spread string must be escaped as text.
func TestRenderAttributesSrcdoc(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "srcdoc strings are escaped as text",
			value:    `<script>alert(1)</script>`,
			expected: ` srcdoc="&amp;lt;script&amp;gt;alert(1)&amp;lt;/script&amp;gt;"`,
		},
		{
			name:     "srcdoc components are rendered as HTML",
			value:    templ.Raw(`<p class="a">trusted</p>`),
			expected: ` srcdoc="&lt;p class=&#34;a&#34;&gt;trusted&lt;/p&gt;"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RenderAttributes(context.Background(), &buf, "iframe", templ.Attributes{"srcdoc": tt.value}); err != nil {
				t.Fatalf("RenderAttributes failed: %v", err)
			}
			if actual := buf.String(); actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
	t.Run("components are not rendered in attributes other than srcdoc", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RenderAttributes(context.Background(), &buf, "div", templ.Attributes{"title": templ.Raw("<p>x</p>")}); err != nil {
			t.Fatalf("RenderAttributes failed: %v", err)
		}
		if actual := buf.String(); actual != "" {
			t.Errorf("expected no attributes, got %q", actual)
		}
	})
}

// TestRenderAttributesOnElements covers GHSA-94c2-xg24-pwhv: spread URL
// attributes are sanitized on any element, as in html/template, while SVG
// animation values are only sanitized on the animation elements.
func TestRenderAttributesOnElements(t *testing.T) {
	attributes := templ.Attributes{
		"action": "javascript:alert(1)",
		"data":   "javascript:alert(2)",
		"src":    "javascript:alert(3)",
		"to":     "javascript:alert(4)",
	}
	tests := []struct {
		name        string
		elementName string
		expected    string
	}{
		{
			name:        "URL attributes are sanitized on a form",
			elementName: "form",
			expected:    ` action="about:invalid#TemplFailedSanitizationURL" data="about:invalid#TemplFailedSanitizationURL" src="about:invalid#TemplFailedSanitizationURL" to="javascript:alert(4)"`,
		},
		{
			name:        "URL attributes are sanitized on custom elements",
			elementName: "turbo-stream",
			expected:    ` action="about:invalid#TemplFailedSanitizationURL" data="about:invalid#TemplFailedSanitizationURL" src="about:invalid#TemplFailedSanitizationURL" to="javascript:alert(4)"`,
		},
		{
			name:        "URL attributes are sanitized on an img",
			elementName: "img",
			expected:    ` action="about:invalid#TemplFailedSanitizationURL" data="about:invalid#TemplFailedSanitizationURL" src="about:invalid#TemplFailedSanitizationURL" to="javascript:alert(4)"`,
		},
		{
			name:        "animation values are sanitized on a set",
			elementName: "set",
			expected:    ` action="about:invalid#TemplFailedSanitizationURL" data="about:invalid#TemplFailedSanitizationURL" src="about:invalid#TemplFailedSanitizationURL" to="about:invalid#TemplFailedSanitizationURL"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RenderAttributes(context.Background(), &buf, tt.elementName, attributes); err != nil {
				t.Fatalf("RenderAttributes failed: %v", err)
			}
			if actual := buf.String(); actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestAttributeExpressionResolveDynamic covers GHSA-94c2-xg24-pwhv: the value of an
// attribute key expression must be sanitized based on the name, because the
// name is only known at render time.
func TestAttributeExpressionResolveDynamic(t *testing.T) {
	tests := []struct {
		name        string
		elementName string
		attrName    string
		value       string
		expected    string
	}{
		{name: "href values are URL sanitized", elementName: "a", attrName: "href", value: "javascript:alert(1)", expected: "about:invalid#TemplFailedSanitizationURL"},
		{name: "mixed case Href values are URL sanitized", elementName: "a", attrName: "Href", value: "javascript:alert(1)", expected: "about:invalid#TemplFailedSanitizationURL"},
		{name: "img src values are URL sanitized, as in html/template", elementName: "img", attrName: "src", value: "data:image/gif;base64,R0lGODlhAQABAAAAACw=", expected: "about:invalid#TemplFailedSanitizationURL"},
		{name: "data-url values are URL sanitized, as in html/template", elementName: "div", attrName: "data-url", value: "javascript:alert(1)", expected: "about:invalid#TemplFailedSanitizationURL"},
		{name: "string event handlers are replaced", elementName: "button", attrName: "onclick", value: "alert(1)", expected: templ.FailedSanitizationJS},
		{name: "string htmx event handlers are replaced", elementName: "button", attrName: "hx-on-click", value: "alert(1)", expected: templ.FailedSanitizationJS},
		{name: "srcdoc values are escaped as text", elementName: "iframe", attrName: "srcdoc", value: "<b>", expected: "&amp;lt;b&amp;gt;"},
		{name: "other values are HTML-escaped", elementName: "div", attrName: "title", value: `"x"`, expected: "&#34;x&#34;"},
		{name: "values of invalid attribute names are not sanitized as event handlers", elementName: "div", attrName: " onclick", value: "x", expected: "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAttributeExpression(tt.value).ResolveDynamic(context.Background(), tt.elementName, tt.attrName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
	t.Run("ComponentScript event handlers are rendered", func(t *testing.T) {
		got, err := NewAttributeExpression(templ.JSFuncCall("alert", "hello")).ResolveDynamic(context.Background(), "button", "onclick")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expected := "alert(&#34;hello&#34;)"; got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})
	t.Run("SafeURL values bypass URL sanitization", func(t *testing.T) {
		got, err := NewAttributeExpression(templ.SafeURL("javascript:alert(1)")).ResolveDynamic(context.Background(), "a", "href")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expected := "javascript:alert(1)"; got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})
}

// TestSrcdocExpressionResolve covers GHSA-94c2-xg24-pwhv: the browser
// parses the value of an iframe srcdoc attribute as an HTML document in the
// context of the page, so a string must be escaped as text.
func TestSrcdocExpressionResolve(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{name: "strings are escaped as text", value: `<script>alert(1)</script>`, expected: "&amp;lt;script&amp;gt;alert(1)&amp;lt;/script&amp;gt;"},
		{name: "integers are formatted", value: 42, expected: "42"},
		{name: "components are rendered as HTML", value: templ.Raw(`<p class="a">x</p>`), expected: "&lt;p class=&#34;a&#34;&gt;x&lt;/p&gt;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSrcdocExpression(tt.value).Resolve(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
	t.Run("unsupported types return an error, rather than being formatted", func(t *testing.T) {
		if _, err := NewSrcdocExpression(new("<b>")).Resolve(context.Background()); err == nil {
			t.Error("expected an error for a string pointer, got nil")
		}
	})
	t.Run("component errors are returned", func(t *testing.T) {
		expectedErr := errors.New("expected")
		_, err := NewSrcdocExpression(templ.Raw("x", expectedErr)).Resolve(context.Background())
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestSanitizeAttributeName covers GHSA-94c2-xg24-pwhv: an attribute name that
// is only known at render time must not be able to add attributes, e.g.
// "x onmouseover=alert(1)".
func TestSanitizeAttributeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "valid names are unchanged", input: "data-test-id", expected: "data-test-id"},
		{name: "names with spaces are replaced", input: "x onclick", expected: templ.FailedSanitizationAttributeName},
		{name: "names with equals signs are replaced", input: "x=y", expected: templ.FailedSanitizationAttributeName},
		{name: "empty names are replaced", input: "", expected: templ.FailedSanitizationAttributeName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeAttributeName(tt.input); got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestAttributeExpressionResolveURL covers GHSA-94c2-xg24-pwhv: values of all types
// in URL attributes must be sanitized, without requiring a string type, so that
// extending the list of URL attributes does not stop templates from compiling.
func TestAttributeExpressionResolveURL(t *testing.T) {
	type CustomString string
	tests := []struct {
		name     string
		run      func() (string, error)
		expected string
	}{
		{
			name: "javascript URLs are sanitized",
			run: func() (string, error) {
				return NewAttributeExpression("javascript:alert(1)").ResolveURL(context.Background())
			},
			expected: string(templ.FailedSanitizationURL),
		},
		{
			name: "mixed case javascript URLs are sanitized",
			run: func() (string, error) {
				return NewAttributeExpression("JaVaScRiPt:alert(1)").ResolveURL(context.Background())
			},
			expected: string(templ.FailedSanitizationURL),
		},
		{
			name: "javascript URLs with whitespace in the scheme are sanitized",
			run: func() (string, error) {
				return NewAttributeExpression("java\tscript:alert(1)").ResolveURL(context.Background())
			},
			expected: string(templ.FailedSanitizationURL),
		},
		{
			name: "data URLs are sanitized",
			run: func() (string, error) {
				return NewAttributeExpression("data:text/html,<script>alert(1)</script>").ResolveURL(context.Background())
			},
			expected: string(templ.FailedSanitizationURL),
		},
		{
			name: "custom string types are sanitized",
			run: func() (string, error) {
				return NewAttributeExpression(CustomString("javascript:alert(1)")).ResolveURL(context.Background())
			},
			expected: string(templ.FailedSanitizationURL),
		},
		{
			name: "safe URLs are HTML-escaped",
			run: func() (string, error) {
				return NewAttributeExpression(`https://example.com/?a=1&b="2"`).ResolveURL(context.Background())
			},
			expected: `https://example.com/?a=1&amp;b=&#34;2&#34;`,
		},
		{
			name: "SafeURL values bypass sanitization",
			run: func() (string, error) {
				return NewAttributeExpression(templ.SafeURL("data:image/gif;base64,R0lGODlhAQABAAAAACw=")).ResolveURL(context.Background())
			},
			expected: "data:image/gif;base64,R0lGODlhAQABAAAAACw=",
		},
		{
			name:     "integer values are formatted",
			run:      func() (string, error) { return NewAttributeExpression(42).ResolveURL(context.Background()) },
			expected: "42",
		},
		{
			name:     "boolean values are formatted",
			run:      func() (string, error) { return NewAttributeExpression(true).ResolveURL(context.Background()) },
			expected: "true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
	t.Run("errors are returned", func(t *testing.T) {
		expectedErr := errors.New("expected")
		_, err := NewAttributeExpression("https://example.com", expectedErr).ResolveURL(context.Background())
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestAttributeExpressionResolveAnimationValue covers GHSA-94c2-xg24-pwhv: SVG animation values
// can set the href of the parent element, and contain a semicolon separated
// list, so each item must be sanitized.
func TestAttributeExpressionResolveAnimationValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "javascript URLs are sanitized", input: "javascript:alert(1)", expected: string(templ.FailedSanitizationURL)},
		{name: "javascript URLs in a list are sanitized", input: "https://example.com;javascript:alert(1)", expected: "https://example.com;" + string(templ.FailedSanitizationURL)},
		{name: "javascript URLs with leading whitespace in a list are sanitized", input: "0; javascript:alert(1)", expected: "0;" + string(templ.FailedSanitizationURL)},
		{name: "numeric lists are unchanged", input: "0;0.5; 1", expected: "0;0.5; 1"},
		{name: "colors are unchanged", input: "red;rgb(0, 0, 255)", expected: "red;rgb(0, 0, 255)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAttributeExpression(tt.input).ResolveAnimationValue(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
	t.Run("SafeURL values bypass sanitization", func(t *testing.T) {
		got, err := NewAttributeExpression(templ.SafeURL("javascript:alert(1)")).ResolveAnimationValue(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "javascript:alert(1)" {
			t.Errorf("got %q, expected %q", got, "javascript:alert(1)")
		}
	})
}

type customString string

type stringerString string

func (s stringerString) String() string {
	return "stringer:" + string(s)
}

type customBool bool

type customInt int

func TestFormatAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{name: "strings are unchanged", value: "a&b", expected: "a&b"},
		{name: "custom string types are formatted as their underlying string", value: customString("value"), expected: "value"},
		{name: "SafeURL values are formatted as their underlying string", value: templ.SafeURL("https://example.com"), expected: "https://example.com"},
		{name: "types with a String method are formatted with it, in the same way as fmt.Sprint", value: stringerString("value"), expected: "stringer:value"},
		{name: "booleans are formatted", value: true, expected: "true"},
		{name: "custom boolean types are formatted", value: customBool(false), expected: "false"},
		{name: "integers are formatted", value: int64(-42), expected: "-42"},
		{name: "custom integer types are formatted", value: customInt(42), expected: "42"},
		{name: "unsigned integers are formatted", value: uint8(8), expected: "8"},
		{name: "floats are formatted", value: 1.5, expected: "1.5"},
		{name: "complex numbers are formatted", value: complex(1, 2), expected: "(1+2i)"},
		{name: "ComponentScript values are formatted as the call", value: templ.JSFuncCall("alert", "x"), expected: "alert(&#34;x&#34;)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := formatAttributeValue(tt.value)
			if !ok {
				t.Fatalf("expected %T to be supported", tt.value)
			}
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
	unsupported := []struct {
		name  string
		value any
	}{
		{name: "string pointers are not supported", value: new("value")},
		{name: "structs are not supported", value: struct{}{}},
		{name: "slices are not supported", value: []string{"value"}},
		{name: "nil is not supported", value: nil},
	}
	for _, tt := range unsupported {
		t.Run(tt.name, func(t *testing.T) {
			if actual, ok := formatAttributeValue(tt.value); ok {
				t.Errorf("expected %T to be unsupported, got %q", tt.value, actual)
			}
		})
	}
}

func TestAttributeExpressionResolve(t *testing.T) {
	tests := []struct {
		name     string
		run      func(context.Context) (string, error)
		expected string
	}{
		{
			name:     "strings are HTML-escaped",
			run:      NewAttributeExpression(`<script>alert("xss")</script>`).Resolve,
			expected: `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`,
		},
		{
			name:     "integers are formatted",
			run:      NewAttributeExpression(42).Resolve,
			expected: "42",
		},
		{
			name:     "booleans are formatted",
			run:      NewAttributeExpression(true).Resolve,
			expected: "true",
		},
		{
			name:     "javascript URLs in attributes that aren't URLs are unchanged",
			run:      NewAttributeExpression("javascript:alert(1)").Resolve,
			expected: "javascript:alert(1)",
		},
		{
			name:     "ComponentScript values are converted to their function call",
			run:      NewAttributeExpression(templ.ComponentScript{Call: `alert(&#34;Hello&#34;)`}).Resolve,
			expected: `alert(&#34;Hello&#34;)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
	t.Run("expression errors are returned", func(t *testing.T) {
		expectedErr := errors.New("expected")
		if _, err := NewAttributeExpression("value", expectedErr).Resolve(context.Background()); !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
