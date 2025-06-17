package parser

import (
	"testing"

	"github.com/a-h/parse"
)

func TestJSXComponentParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *JSXComponentElement
		wantErr  bool
	}{
		{
			name:  "jsx: self-closing component no attributes",
			input: `<Button />`,
			expected: &JSXComponentElement{
				Name:        "Button",
				SelfClosing: true,
			},
		},
		{
			name:  "jsx: self-closing component with expression attribute",
			input: `<Button text={variable} />`,
			expected: &JSXComponentElement{
				Name:        "Button",
				SelfClosing: true,
				Attributes: []Attribute{
					&ExpressionAttribute{
						Key: ConstantAttributeKey{Name: "text"},
						Expression: Expression{
							Value: "variable",
						},
					},
				},
			},
		},
		{
			name:  "jsx: self-closing component with string attribute",
			input: `<Button text="Click me" />`,
			expected: &JSXComponentElement{
				Name:        "Button",
				SelfClosing: true,
				Attributes: []Attribute{
					&ConstantAttribute{
						Key:   ConstantAttributeKey{Name: "text"},
						Value: "Click me",
					},
				},
			},
		},
		{
			name:  "jsx: self-closing component with multiple attributes",
			input: `<DData term="Name" detail="Tom Cook" />`,
			expected: &JSXComponentElement{
				Name:        "DData",
				SelfClosing: true,
				Attributes: []Attribute{
					&ConstantAttribute{
						Key:   ConstantAttributeKey{Name: "term"},
						Value: "Name",
					},
					&ConstantAttribute{
						Key:   ConstantAttributeKey{Name: "detail"},
						Value: "Tom Cook",
					},
				},
			},
		},
		{
			name:  "jsx: component with children",
			input: `<DList><div>Child content</div></DList>`,
			expected: &JSXComponentElement{
				Name:        "DList",
				SelfClosing: false,
				Children: []Node{
					&Element{
						Name: "div",
						Children: []Node{
							&Text{Value: "Child content"},
						},
					},
				},
			},
		},
		{
			name:  "jsx: component with package prefix",
			input: `<components.Button text="Click" />`,
			expected: &JSXComponentElement{
				Name:        "components.Button",
				SelfClosing: true,
				Attributes: []Attribute{
					&ConstantAttribute{
						Key:   ConstantAttributeKey{Name: "text"},
						Value: "Click",
					},
				},
			},
		},
		{
			name:    "jsx: not a component (lowercase)",
			input:   `<div>content</div>`,
			wantErr: false, // Should not be parsed by JSX parser, falls through to element parser
		},
		{
			name:  "jsx: component with expression attribute",
			input: `<Button text={variable} />`,
			expected: &JSXComponentElement{
				Name:        "Button",
				SelfClosing: true,
				Attributes: []Attribute{
					&ExpressionAttribute{
						Key: ConstantAttributeKey{Name: "text"},
						Expression: Expression{
							Value: "variable",
						},
					},
				},
			},
		},
		{
			name:  "jsx: component with boolean attribute",
			input: `<Button disabled />`,
			expected: &JSXComponentElement{
				Name:        "Button",
				SelfClosing: true,
				Attributes: []Attribute{
					&BoolConstantAttribute{
						Key: ConstantAttributeKey{Name: "disabled"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := parse.NewInput(tt.input)
			node, matched, err := jsxComponent.Parse(pi)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expected == nil {
				// Test expects no match (e.g., lowercase element names)
				if matched {
					t.Error("expected no match but got match")
				}
				return
			}

			if !matched {
				t.Error("expected match but got no match")
				return
			}

			result, ok := node.(*JSXComponentElement)
			if !ok {
				t.Errorf("expected *JSXComponentElement but got %T", node)
				return
			}

			// Compare component name
			if result.Name != tt.expected.Name {
				t.Errorf("name mismatch:\nexpected: %q\ngot:      %q",
					tt.expected.Name, result.Name)
			}

			// Compare self-closing flag
			if result.SelfClosing != tt.expected.SelfClosing {
				t.Errorf("self-closing mismatch:\nexpected: %v\ngot:      %v",
					tt.expected.SelfClosing, result.SelfClosing)
			}

			// Compare attribute count
			if len(result.Attributes) != len(tt.expected.Attributes) {
				t.Errorf("attribute count mismatch:\nexpected: %d\ngot:      %d",
					len(tt.expected.Attributes), len(result.Attributes))
			}

			// For tests with children, check that we have some children
			if len(tt.expected.Children) > 0 && len(result.Children) == 0 {
				t.Error("expected children but got none")
			}
		})
	}
}

func TestJSXComponentParserIntegration(t *testing.T) {
	// Test that JSX components work within templates
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "jsx in template",
			input: `package main

templ Page() {
	<DData term="Name" detail="Tom Cook" />
}`,
		},
		{
			name: "jsx mixed with regular elements",
			input: `package main

templ Page() {
	<div>
		<h1>Title</h1>
		<Button text="Click me" />
	</div>
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseString(tt.input)
			if err != nil {
				t.Errorf("failed to parse template with JSX: %v", err)
			}
		})
	}
}
