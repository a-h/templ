package generator

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAttributeResolverFormatsTheMethodCall(t *testing.T) {
	tests := []struct {
		name     string
		resolver attributeResolver
		expected string
	}{
		{
			name:     "a resolver without arguments passes the context",
			resolver: attributeResolverURL,
			expected: "ResolveURL(ctx)",
		},
		{
			name:     "a dynamic resolver passes the element writer, the element name, and the attribute name",
			resolver: newDynamicAttributeResolver("button", "templ_7745c5c3_Var1"),
			expected: `ResolveDynamic(ctx, templ_7745c5c3_ElementWriter, "button", templ_7745c5c3_Var1)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.expected, tt.resolver.FormatMethodCall()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
