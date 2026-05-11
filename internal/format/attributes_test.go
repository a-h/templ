package format

import (
	"testing"

	parser "github.com/a-h/templ/parser/v2"
)

func TestAttributes(t *testing.T) {
	tests := []struct {
		name     string
		children []parser.Node
		command  string
		check    func(t *testing.T, children []parser.Node)
	}{
		{
			name:     "no children returns nil without calling prettier",
			children: nil,
			command:  "nonexistent-command-that-would-fail",
		},
		{
			name: "no constant attributes returns nil without calling prettier",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ExpressionAttribute{},
					},
				},
			},
			command: "nonexistent-command-that-would-fail",
		},
		{
			name: "expression attribute keys are skipped",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{Key: parser.ExpressionAttributeKey{}, Value: "foo"},
					},
				},
			},
			command: "nonexistent-command-that-would-fail",
		},
		{
			name: "passthrough with cat preserves values",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "class"},
							Value: "text-lg font-medium",
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.Value != "text-lg font-medium" {
					t.Errorf("got %q, expected %q", ca.Value, "text-lg font-medium")
				}
			},
		},
		{
			name: "empty attribute value survives round-trip",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "name"},
							Value: "",
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.Value != "" {
					t.Errorf("got %q, expected empty string", ca.Value)
				}
			},
		},
		{
			name: "single quote flag is re-evaluated when value contains double quote",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:         parser.ConstantAttributeKey{Name: "title"},
							Value:       `He said "hello"`,
							SingleQuote: false,
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if !ca.SingleQuote {
					t.Error("expected SingleQuote to be true for value containing double quote")
				}
			},
		},
		{
			name: "single quote flag is cleared when value does not contain double quote",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:         parser.ConstantAttributeKey{Name: "class"},
							Value:       "no-quotes-here",
							SingleQuote: true,
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.SingleQuote {
					t.Error("expected SingleQuote to be false for value without double quote")
				}
			},
		},
		{
			name: "HTML special characters round-trip correctly",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "data-value"},
							Value: `<script>&amp;`,
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.Value != `<script>&amp;` {
					t.Errorf("got %q, expected %q", ca.Value, `<script>&amp;`)
				}
			},
		},
		{
			name: "multiple elements are correlated correctly",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "class"},
							Value: "first",
						},
					},
				},
				&parser.Element{
					Name: "span",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "class"},
							Value: "second",
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				first := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				second := children[1].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if first.Value != "first" {
					t.Errorf("first element: got %q, expected %q", first.Value, "first")
				}
				if second.Value != "second" {
					t.Errorf("second element: got %q, expected %q", second.Value, "second")
				}
			},
		},
		{
			name: "conditional branches with duplicate keys are formatted independently",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConditionalAttribute{
							Then: []parser.Attribute{
								&parser.ConstantAttribute{
									Key:   parser.ConstantAttributeKey{Name: "class"},
									Value: "then-class",
								},
							},
							Else: []parser.Attribute{
								&parser.ConstantAttribute{
									Key:   parser.ConstantAttributeKey{Name: "class"},
									Value: "else-class",
								},
							},
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				cond := children[0].(*parser.Element).Attributes[0].(*parser.ConditionalAttribute)
				thenAttr := cond.Then[0].(*parser.ConstantAttribute)
				elseAttr := cond.Else[0].(*parser.ConstantAttribute)
				if thenAttr.Value != "then-class" {
					t.Errorf("then branch: got %q, expected %q", thenAttr.Value, "then-class")
				}
				if elseAttr.Value != "else-class" {
					t.Errorf("else branch: got %q, expected %q", elseAttr.Value, "else-class")
				}
			},
		},
		{
			name: "nested elements are collected via visitor",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "class"},
							Value: "outer",
						},
					},
					Children: []parser.Node{
						&parser.Element{
							Name: "span",
							Attributes: []parser.Attribute{
								&parser.ConstantAttribute{
									Key:   parser.ConstantAttributeKey{Name: "class"},
									Value: "inner",
								},
							},
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				outer := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				inner := children[0].(*parser.Element).Children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if outer.Value != "outer" {
					t.Errorf("outer: got %q, expected %q", outer.Value, "outer")
				}
				if inner.Value != "inner" {
					t.Errorf("inner: got %q, expected %q", inner.Value, "inner")
				}
			},
		},
		{
			name: "elements inside control flow are collected via visitor",
			children: []parser.Node{
				&parser.IfExpression{
					Then: []parser.Node{
						&parser.Element{
							Name: "div",
							Attributes: []parser.Attribute{
								&parser.ConstantAttribute{
									Key:   parser.ConstantAttributeKey{Name: "class"},
									Value: "conditional",
								},
							},
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.IfExpression).Then[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.Value != "conditional" {
					t.Errorf("got %q, expected %q", ca.Value, "conditional")
				}
			},
		},
		{
			name: "mixed-case attribute keys survive HTML parser lowercasing",
			children: []parser.Node{
				&parser.Element{
					Name: "select",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "optionC"},
							Value: "test",
						},
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "onMouseover"},
							Value: "handler()",
						},
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "CLASS"},
							Value: "my-class",
						},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				attrs := children[0].(*parser.Element).Attributes
				ca0 := attrs[0].(*parser.ConstantAttribute)
				if ca0.Value != "test" {
					t.Errorf("optionC: got %q, expected %q", ca0.Value, "test")
				}
				ca1 := attrs[1].(*parser.ConstantAttribute)
				if ca1.Value != "handler()" {
					t.Errorf("onMouseover: got %q, expected %q", ca1.Value, "handler()")
				}
				ca2 := attrs[2].(*parser.ConstantAttribute)
				if ca2.Value != "my-class" {
					t.Errorf("CLASS: got %q, expected %q", ca2.Value, "my-class")
				}
			},
		},
		{
			name: "mixed attribute types only collect constant attributes",
			children: []parser.Node{
				&parser.Element{
					Name: "div",
					Attributes: []parser.Attribute{
						&parser.ConstantAttribute{
							Key:   parser.ConstantAttributeKey{Name: "class"},
							Value: "keep",
						},
						&parser.ExpressionAttribute{},
						&parser.BoolConstantAttribute{},
						&parser.BoolExpressionAttribute{},
						&parser.SpreadAttributes{},
					},
				},
			},
			command: "cat",
			check: func(t *testing.T, children []parser.Node) {
				t.Helper()
				ca := children[0].(*parser.Element).Attributes[0].(*parser.ConstantAttribute)
				if ca.Value != "keep" {
					t.Errorf("got %q, expected %q", ca.Value, "keep")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Attributes(tt.children, tt.command)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, tt.children)
			}
		})
	}
}
