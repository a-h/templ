package templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Note: use of the RawEventHandler and JSFuncCall in ExpressionAttributes is tested in the parser package.

func TestJSUnsafeFuncCall(t *testing.T) {
	tests := []struct {
		name     string
		js       string
		expected ComponentScript
	}{
		{
			name: "alert",
			js:   "alert('hello')",
			expected: ComponentScript{
				Name:     "jsUnsafeFuncCall_bc8b29d9abedc43cb4d79ec0af23be8c4255a4b76691aecf23ba3b0b8ab90011",
				Function: "",
				// Note that the Call field is attribute encoded.
				Call: "alert(&#39;hello&#39;)",
				// Whereas the CallInline field is what you would see inside a <script> tag.
				CallInline: "alert('hello')",
			},
		},
		{
			name: "immediately executed function",
			js:   "(function(x) { x < 3 ? alert('hello less than 3') : alert('more than 3'); })(2);",
			expected: ComponentScript{
				Name:       "jsUnsafeFuncCall_e4c24908f83227fd10c1a984fe8f99e15bfb3f195985af517253f6a72ec9106b",
				Function:   "",
				Call:       "(function(x) { x &lt; 3 ? alert(&#39;hello less than 3&#39;) : alert(&#39;more than 3&#39;); })(2);",
				CallInline: "(function(x) { x < 3 ? alert('hello less than 3') : alert('more than 3'); })(2);",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := JSUnsafeFuncCall(tt.js)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestJSFuncCall(t *testing.T) {
	tests := []struct {
		name                    string
		functionName            string
		args                    []any
		expected                ComponentScript
		expectedComponentOutput string
	}{
		{
			name:         "no arguments are supported",
			functionName: "doSomething",
			args:         nil,
			expected: ComponentScript{
				Name:       "jsFuncCall_6e742483001f8d3c67652945c597b5a2025a7411cb9d1bae2f9e160bebfeb4c6",
				Function:   "",
				Call:       "doSomething()",
				CallInline: "doSomething()",
			},
			expectedComponentOutput: `<script>doSomething()</script>`,
		},
		{
			name:         "single argument is supported",
			functionName: "alert",
			args:         []any{"hello"},
			expected: ComponentScript{
				Name:       "jsFuncCall_92df7244f17dc5bfc41dfd02043df695e4664f8bf42c265a46d79b32b97693d0",
				Function:   "",
				Call:       "alert(&#34;hello&#34;)",
				CallInline: `alert("hello")`,
			},
			expectedComponentOutput: `<script>alert("hello")</script>`,
		},
		{
			name:         "multiple arguments are supported",
			functionName: "console.log",
			args:         []any{"hello", "world"},
			expected: ComponentScript{
				Name:       "jsFuncCall_2b3416c14fc2700d01e0013e7b7076bb8dd5f3126d19e2e801de409163e3960c",
				Function:   "",
				Call:       "console.log(&#34;hello&#34;,&#34;world&#34;)",
				CallInline: `console.log("hello","world")`,
			},
			expectedComponentOutput: `<script>console.log("hello","world")</script>`,
		},
		{
			name:         "attribute injection fails",
			functionName: `" onmouseover="alert('hello')`,
			args:         nil,
			expected: ComponentScript{
				Name:       "jsFuncCall_e56d1214f3b4fbf27406f209e3f4a58c2842fa2760b6d83da5ee72e04c89f913",
				Function:   "",
				Call:       "__templ_invalid_js_function_name()",
				CallInline: "__templ_invalid_js_function_name()",
			},
			expectedComponentOutput: `<script>__templ_invalid_js_function_name()</script>`,
		},
		{
			name:         "closing the script and injecting HTML fails",
			functionName: `</script><div>Hello</div><script>`,
			args:         nil,
			expected: ComponentScript{
				Name:       "jsFuncCall_e56d1214f3b4fbf27406f209e3f4a58c2842fa2760b6d83da5ee72e04c89f913",
				Function:   "",
				Call:       "__templ_invalid_js_function_name()",
				CallInline: "__templ_invalid_js_function_name()",
			},
			expectedComponentOutput: `<script>__templ_invalid_js_function_name()</script>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test creation.
			actual := JSFuncCall(tt.functionName, tt.args...)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}

			// Test rendering.
			buf := new(bytes.Buffer)
			err := actual.Render(context.Background(), buf)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.expectedComponentOutput, buf.String()); diff != "" {
				t.Error(diff)
			}

		})
	}
}

func TestJSFunctionNameRegexp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "console.log",
			expected: true,
		},
		{
			input:    "alert",
			expected: true,
		},
		{
			input:    "console.log('hello')",
			expected: false,
		},
		{
			input:    "</script><div>Hello</div><script>",
			expected: false,
		},
		{
			input:    `" onmouseover="alert('hello')`,
			expected: false,
		},
		{
			input:    "(new Date()).getTime",
			expected: false,
		},
		{
			input:    "expressionThatReturnsAFunction()",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			actual := jsFunctionName.MatchString(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}
