package parser

import (
	"testing"

	"github.com/a-h/parse"
)

func TestInlineComponentAttributeParser(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectMatch    bool
		expectError    bool
		expectedKey    string
		expectedResult string // Expected content or error description
	}{
		// Should NOT match - regular expressions
		{
			name:        "simple string literal",
			input:       `data-attr={ "raw" }`,
			expectMatch: false,
		},
		{
			name:        "function call",
			input:       `data-attr={ funcWithNoError() }`,
			expectMatch: false,
		},
		{
			name:        "function call with error",
			input:       `data-attr={ funcWithError(err) }`,
			expectMatch: false,
		},
		{
			name:        "string with angle brackets",
			input:       `data-contents={ "something with <tag> inside" }`,
			expectMatch: false,
		},
		{
			name:        "backtick string with HTML",
			input:       "data-contents={ `something with <div>tags</div> inside` }",
			expectMatch: false,
		},
		{
			name:        "single quote string with HTML",
			input:       `data-contents={ 'something with <span>content</span> inside' }`,
			expectMatch: false,
		},
		{
			name:        "variable reference",
			input:       `value={ someVariable }`,
			expectMatch: false,
		},
		{
			name:        "complex expression",
			input:       `href={ templ.URL("mailto: " + p.Email) }`,
			expectMatch: false,
		},
		
		// Should match - actual inline components
		{
			name:           "simple div element",
			input:          `child={ <div>content</div> }`,
			expectMatch:    true,
			expectedKey:    "child",
			expectedResult: "div element with text content",
		},
		{
			name:           "self-closing element",
			input:          `icon={ <img src="icon.png" /> }`,
			expectMatch:    true,
			expectedKey:    "icon",
			expectedResult: "img element",
		},
		{
			name:           "multiple elements",
			input:          `content={ <div><p>Hello</p><p>World</p></div> }`,
			expectMatch:    true,
			expectedKey:    "content",
			expectedResult: "nested elements",
		},
		{
			name:           "component element",
			input:          `child={ <Button title="Click me" /> }`,
			expectMatch:    true,
			expectedKey:    "child",
			expectedResult: "component element",
		},
		
		// Edge cases
		{
			name:        "empty braces",
			input:       `empty={ }`,
			expectMatch: false,
		},
		{
			name:        "whitespace only",
			input:       `whitespace={   }`,
			expectMatch: false,
		},
		{
			name:        "nested braces in expression",
			input:       `data={ map[string]string{"key": "value"} }`,
			expectMatch: false,
		},
		{
			name:           "HTML with nested braces",
			input:          `content={ <div data-value={ getValue() }>Text</div> }`,
			expectMatch:    true,
			expectedKey:    "content",
			expectedResult: "HTML with nested expressions",
		},
		
		// Additional edge cases
		{
			name:        "Go expression with comparison operators",
			input:       `condition={ a < b && c > d }`,
			expectMatch: false,
		},
		{
			name:        "string literal with escaped quotes",
			input:       `data={ "text with \"quotes\" and <tag>" }`,
			expectMatch: false,
		},
		{
			name:        "multiline string with HTML",
			input:       "data={ `multiline\nstring with <div>tags</div>` }",
			expectMatch: false,
		},
		{
			name:        "complex Go expression with angle brackets",
			input:       `data={ fmt.Sprintf("value: %d", getValue() < 10 ? 1 : 2) }`,
			expectMatch: false,
		},
		{
			name:           "inline component with Go expression attribute",
			input:          `child={ <div data-value={ getValue() }>content</div> }`,
			expectMatch:    true,
			expectedKey:    "child",
			expectedResult: "div with expression attribute",
		},
		
		// Error cases
		{
			name:        "unclosed braces",
			input:       `broken={ <div>content`,
			expectMatch: false,
		},
		{
			name:        "malformed HTML in inline component",
			input:       `bad={ <div><span>unclosed }`,
			expectMatch: true,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			
			attr, matched, err := inlineComponentAttributeParser.Parse(input)
			
			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			// Check match expectation
			if matched != tt.expectMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.expectMatch, matched)
				return
			}
			
			if !tt.expectMatch {
				// For non-matches, we expect no error and nil result
				if err != nil {
					t.Errorf("Unexpected error for non-match: %v", err)
				}
				if attr != nil {
					t.Errorf("Expected nil attribute for non-match, got %+v", attr)
				}
				return
			}
			
			// For matches, verify the results
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if attr == nil {
				t.Errorf("Expected non-nil attribute for match")
				return
			}
			
			if attr.Key.String() != tt.expectedKey {
				t.Errorf("Expected key=%q, got key=%q", tt.expectedKey, attr.Key.String())
			}
			
			// Verify that children were parsed (basic check)
			if len(attr.Children) == 0 {
				t.Errorf("Expected parsed children, got empty children")
			}
		})
	}
}

func TestInlineComponentAttributeParserIntegration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectTypes []string // Expected attribute types in order
	}{
		{
			name:        "mixed attributes with inline component",
			input:       `class="test" data-value={ getValue() } child={ <div>content</div> }`,
			expectTypes: []string{"ConstantAttribute", "ExpressionAttribute", "InlineComponentAttribute"},
		},
		{
			name:        "all expression attributes",
			input:       `data-a={ "string" } data-b={ func() } data-c={ "text with <tags>" }`,
			expectTypes: []string{"ExpressionAttribute", "ExpressionAttribute", "ExpressionAttribute"},
		},
		{
			name:        "all inline components",
			input:       `child1={ <div>1</div> } child2={ <span>2</span> }`,
			expectTypes: []string{"InlineComponentAttribute", "InlineComponentAttribute"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			
			parser := attributesParser{}
			attrs, matched, err := parser.Parse(input)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !matched {
				t.Errorf("Expected to match attributes")
				return
			}
			
			if len(attrs) != len(tt.expectTypes) {
				t.Errorf("Expected %d attributes, got %d", len(tt.expectTypes), len(attrs))
				return
			}
			
			for i, expectedType := range tt.expectTypes {
				actualType := ""
				switch attrs[i].(type) {
				case *ConstantAttribute:
					actualType = "ConstantAttribute"
				case *ExpressionAttribute:
					actualType = "ExpressionAttribute" 
				case *InlineComponentAttribute:
					actualType = "InlineComponentAttribute"
				case *BoolConstantAttribute:
					actualType = "BoolConstantAttribute"
				case *BoolExpressionAttribute:
					actualType = "BoolExpressionAttribute"
				default:
					actualType = "Unknown"
				}
				
				if actualType != expectedType {
					t.Errorf("Attribute %d: expected type %s, got type %s", i, expectedType, actualType)
				}
			}
		})
	}
}

// Benchmark the inline component detection logic
func BenchmarkInlineComponentAttributeParser(b *testing.B) {
	testCases := []string{
		`data-attr={ "simple string" }`,
		`data-attr={ someFunction() }`,
		`data-contents={ "string with <tags>" }`,
		`child={ <div>simple element</div> }`,
		`complex={ <div><p>nested</p><span>elements</span></div> }`,
	}
	
	for _, testCase := range testCases {
		b.Run(testCase, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				input := parse.NewInput(testCase)
				_, _, _ = inlineComponentAttributeParser.Parse(input)
			}
		})
	}
}