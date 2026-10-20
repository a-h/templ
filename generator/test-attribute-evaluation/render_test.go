package testattributeevaluation

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// TestAttributeExpressionsAreEvaluatedOnceInTheSelectedBranch covers attribute
// expressions that render a script or CSS class before the opening tag of an
// element. Each expression must be evaluated once, so that the rendered
// function is the one that the attribute calls, and only when its branch of a
// conditional attribute is selected.
func TestAttributeExpressionsAreEvaluatedOnceInTheSelectedBranch(t *testing.T) {
	tests := []struct {
		name          string
		render        func(r *recorder) templ.Component
		expectedCalls []string
		expected      string
	}{
		{
			name:          "an event handler is evaluated once",
			render:        func(r *recorder) templ.Component { return renderEventHandler(r) },
			expectedCalls: []string{"handler"},
			expected:      `<script>function handler_1(){}</script><button onclick="handler_1()">button</button>`,
		},
		{
			name:          "an event handler in the then branch is evaluated only when the condition is true",
			render:        func(r *recorder) templ.Component { return renderConditionalEventHandler(r, true) },
			expectedCalls: []string{"then"},
			expected:      `<script>function then_1(){}</script><button onclick="then_1()">button</button>`,
		},
		{
			name:          "an event handler in the else branch is evaluated only when the condition is false",
			render:        func(r *recorder) templ.Component { return renderConditionalEventHandler(r, false) },
			expectedCalls: []string{"else"},
			expected:      `<script>function else_1(){}</script><button onclick="else_1()">button</button>`,
		},
		{
			name:          "a dynamic key value is evaluated once",
			render:        func(r *recorder) templ.Component { return renderDynamicKey(r) },
			expectedCalls: []string{"dynamic"},
			expected:      `<script>function dynamic_1(){}</script><button onclick="dynamic_1()">button</button>`,
		},
		{
			name:          "a dynamic key value in a branch is evaluated when the condition is true",
			render:        func(r *recorder) templ.Component { return renderConditionalDynamicKey(r, true, nil) },
			expectedCalls: []string{"dynamic"},
			expected:      `<script>function dynamic_1(){}</script><button onclick="dynamic_1()">button</button>`,
		},
		{
			name:     "a dynamic key value in a branch is not evaluated when the condition is false",
			render:   func(r *recorder) templ.Component { return renderConditionalDynamicKey(r, false, nil) },
			expected: `<button>button</button>`,
		},
		{
			name: "an error from a dynamic key value in a branch that is not selected does not fail rendering",
			render: func(r *recorder) templ.Component {
				return renderConditionalDynamicKey(r, false, errors.New("unselected"))
			},
			expected: `<button>button</button>`,
		},
		{
			name:          "a spread attribute is evaluated once",
			render:        func(r *recorder) templ.Component { return renderSpread(r) },
			expectedCalls: []string{"spread"},
			expected:      `<script>function spread_1(){}</script><button onclick="spread_1()">button</button>`,
		},
		{
			name:     "a spread attribute in a branch is not evaluated when the condition is false",
			render:   func(r *recorder) templ.Component { return renderConditionalSpread(r, false) },
			expected: `<button>button</button>`,
		},
		{
			name:          "a class is evaluated once",
			render:        func(r *recorder) templ.Component { return renderClass(r) },
			expectedCalls: []string{"class"},
			expected:      `<style type="text/css">.class_1{}</style><div class="class_1">div</div>`,
		},
		{
			name:          "an uppercase class attribute renders its CSS in the same way as a lowercase one",
			render:        func(r *recorder) templ.Component { return renderUppercaseClass(r) },
			expectedCalls: []string{"class"},
			expected:      `<style type="text/css">.class_1{}</style><div CLASS="class_1">div</div>`,
		},
		{
			name:          "a class in the then branch is evaluated only when the condition is true",
			render:        func(r *recorder) templ.Component { return renderConditionalClass(r, true) },
			expectedCalls: []string{"then"},
			expected:      `<style type="text/css">.then_1{}</style><div class="then_1">div</div>`,
		},
		{
			name:          "a class in the else branch is evaluated only when the condition is false",
			render:        func(r *recorder) templ.Component { return renderConditionalClass(r, false) },
			expectedCalls: []string{"else"},
			expected:      `<style type="text/css">.else_1{}</style><div class="else_1">div</div>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			var actual strings.Builder
			if err := tt.render(r).Render(context.Background(), &actual); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(r.Calls, tt.expectedCalls) {
				t.Errorf("expected calls %v, got %v", tt.expectedCalls, r.Calls)
			}
			if actual.String() != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual.String())
			}
		})
	}
}
