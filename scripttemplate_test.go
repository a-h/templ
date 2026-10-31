package templ_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestRenderScriptItems(t *testing.T) {
	s1 := templ.ComponentScript{
		Name:     "s1",
		Function: "function s1() { return 'hello1'; }",
	}
	s2 := templ.ComponentScript{
		Name:     "s2",
		Function: "function s2() { return 'hello2'; }",
	}
	tests := []struct {
		name     string
		toIgnore []templ.ComponentScript
		toRender []templ.ComponentScript
		expected string
	}{
		{
			name:     "if none are ignored, everything is rendered",
			toIgnore: nil,
			toRender: []templ.ComponentScript{s1, s2},
			expected: `<script>` + s1.Function + s2.Function + `</script>`,
		},
		{
			name: "if something outside the expected is ignored, if has no effect",
			toIgnore: []templ.ComponentScript{
				{
					Name:     "s3",
					Function: "function s3() { return 'hello3'; }",
				},
			},
			toRender: []templ.ComponentScript{s1, s2},
			expected: `<script>` + s1.Function + s2.Function + `</script>`,
		},
		{
			name:     "if one is ignored, it's not rendered",
			toIgnore: []templ.ComponentScript{s1},
			toRender: []templ.ComponentScript{s1, s2},
			expected: `<script>` + s2.Function + `</script>`,
		},
		{
			name: "if all are ignored, not even style tags are rendered",
			toIgnore: []templ.ComponentScript{
				s1,
				s2,
				{
					Name:     "s3",
					Function: "function s3() { return 'hello3'; }",
				},
			},
			toRender: []templ.ComponentScript{s1, s2},
			expected: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := new(bytes.Buffer)

			// Render twice, reusing the same context so that there's a memory of which classes have been rendered.
			ctx = templ.InitializeContext(ctx)
			err := templ.RenderScriptItems(ctx, b, tt.toIgnore...)
			if err != nil {
				t.Fatalf("failed to render initial scripts: %v", err)
			}

			// Now render again to check that only the expected classes were rendered.
			b.Reset()
			err = templ.RenderScriptItems(ctx, b, tt.toRender...)
			if err != nil {
				t.Fatalf("failed to render scripts: %v", err)
			}

			if diff := cmp.Diff(tt.expected, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRenderScriptItemsWithNonce(t *testing.T) {
	ctx := templ.WithNonce(context.Background(), `test"nonce`)
	b := new(bytes.Buffer)
	s := templ.ComponentScript{
		Name:     "s1",
		Function: "function s1() { return 'hello1'; }",
	}
	err := templ.RenderScriptItems(ctx, b, s)
	if err != nil {
		t.Fatalf("failed to render scripts: %v", err)
	}
	// Should include the escaped nonce attribute on the <script> tag.
	expected := `<script nonce="test&#34;nonce">` + s.Function + `</script>`
	if diff := cmp.Diff(expected, b.String()); diff != "" {
		t.Error(diff)
	}
	t.Run("rendering a script call with a nonce requires one allocation", func(t *testing.T) {
		// Nonces are base64, so they don't need escaping, which would allocate.
		ctx := templ.WithNonce(context.Background(), "dGVzdG5vbmNl")
		call := templ.JSFuncCall("alert", 1)
		actualAllocs := testing.AllocsPerRun(4, func() {
			if err := call.Render(ctx, io.Discard); err != nil {
				t.Fatalf("failed to render script call: %v", err)
			}
		})
		if actualAllocs > 1 {
			t.Errorf("expected 1 alloc, got %v", actualAllocs)
		}
	})
}

func TestJSExpression(t *testing.T) {
	expected := "myJSFunction(\"StringValue\",123,event,1 + 2)"
	actual := templ.SafeScriptInline("myJSFunction", "StringValue", 123, templ.JSExpression("event"), templ.JSExpression("1 + 2"))

	if actual != expected {
		t.Fatalf("TestJSExpression: expected %q, got %q", expected, actual)
	}
}
