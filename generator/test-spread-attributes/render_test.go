package testspreadattributes

import (
	_ "embed"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	component := BasicTemplate(templ.Attributes{
		// Should render as `bool` as the value is true, and the conditional render is also true.
		"bool": templ.KV(true, true),
		// Should not render, as the conditional render value is false.
		"bool-disabled": templ.KV(true, false),
		// Should render non-nil string values.
		"data-attr": ptr("value"),
		// Should render non-nil boolean values that evaluate to true.
		"data-attr-bool": ptr(true),
		// Should render as `dateId="my-custom-id"`.
		"dateId": "my-custom-id",
		// Should render as `hx-get="/page"`.
		"hx-get": "/page",
		// Should render as `id="test"`.
		"id": "test",
		// Should not render a nil string pointer.
		"key": nilPtr[string](),
		// Should not render a nil boolean value.
		"boolkey": nilPtr[bool](),
		// Should not render, as the attribute value, and the conditional render value is false.
		"no-bool": templ.KV(false, false),
		// Should not render, as the conditional render value is false.
		"no-text": templ.KV("empty", false),
		// Should render as `nonshare`, as the value is true.
		"nonshade": true,
		// Should not render, as the value is false.
		"shade": false,
		// Should render text="lorem" as the value is true.
		"text": templ.KV("lorem", true),
		// Optional attribute based on result of func() bool.
		"optional-from-func-false": func() bool { return false },
		// Optional attribute based on result of func() bool.
		"optional-from-func-true": func() bool { return true },
	})

	diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}

//go:embed expected_ordered_attributes.html
var expectedOrderedAttributes string

func TestOrderedAttributes(t *testing.T) {
	component := BasicTemplateOrdered(templ.OrderedAttributes{
		// Should render as `bool` as the value is true, and the conditional render is also true.
		{Key: "bool", Value: templ.KV(true, true)},
		// Should not render, as the conditional render value is false.
		{Key: "bool-disabled", Value: templ.KV(true, false)},
		// Should render non-nil string values.
		{Key: "data-attr", Value: ptr("value")},
		// Should render non-nil boolean values that evaluate to true.
		{Key: "data-attr-bool", Value: ptr(true)},
		// Should render as `dateId="my-custom-id"`.
		{Key: "dateId", Value: "my-custom-id"},
		// Should render as `hx-get="/page"`.
		{Key: "hx-get", Value: "/page"},
		// Should render as `id="test"`.
		{Key: "id", Value: "test"},
		// Should not render a nil string pointer.
		{Key: "key", Value: nilPtr[string]()},
		// Should not render a nil boolean value.
		{Key: "boolkey", Value: nilPtr[bool]()},
		// Should not render, as the attribute value, and the conditional render value is false.
		{Key: "no-bool", Value: templ.KV(false, false)},
		// Should not render, as the conditional render value is false.
		{Key: "no-text", Value: templ.KV("empty", false)},
		// Should render as `nonshare`, as the value is true.
		{Key: "nonshade", Value: true},
		// Should not render, as the value is false.
		{Key: "shade", Value: false},
		// Should render text="lorem" as the value is true.
		{Key: "text", Value: templ.KV("lorem", true)},
		// Optional attribute based on result of func() bool.
		{Key: "optional-from-func-false", Value: func() bool { return false }},
		// Optional attribute based on result of func() bool.
		{Key: "optional-from-func-true", Value: func() bool { return true }},
	})

	diff, err := htmldiff.Diff(component, expectedOrderedAttributes)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}

func nilPtr[T any]() *T {
	return nil
}

func ptr[T any](x T) *T {
	return &x
}

//go:embed expected_numeric_attributes.html
var expectedNumericAttributes string

func TestNumericAttributeTypes(t *testing.T) {
	t.Parallel()
	component := BasicTemplate(templ.Attributes{
		"int-value":     42,
		"float-value":   3.14,
		"uint-value":    uint(100),
		"int64-value":   int64(9223372036854775807),
		"complex-value": complex(1, 2),
		"string-value":  "text",
		"bool-true":     true,
		"bool-false":    false,
	})

	diff, err := htmldiff.Diff(component, expectedNumericAttributes)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
