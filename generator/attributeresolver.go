package generator

import (
	"strconv"
	"strings"
)

// attributeResolver is the generated code that resolves the value of an
// attribute expression: Constructor wraps the result of the expression, and
// Method, called with the context and then Args, converts it to an
// HTML-escaped string, e.g.
// templruntime.NewAttributeExpression(p.URL).ResolveURL(ctx).
type attributeResolver struct {
	Constructor string
	Method      string
	Args        []string
}

var (
	attributeResolverDefault = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "Resolve",
	}
	attributeResolverURL = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveURL",
	}
	attributeResolverAnimationValue = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveAnimationValue",
	}
	attributeResolverSrcset = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveSrcset",
	}
	attributeResolverSrcdoc = attributeResolver{
		Constructor: "templruntime.NewSrcdocExpression",
		Method:      "Resolve",
	}
)

func newDynamicAttributeResolver(elementName, nameVariable string) attributeResolver {
	return attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveDynamic",
		Args:        []string{elementWriterVariable, strconv.Quote(elementName), nameVariable},
	}
}

// FormatMethodCall returns the Go code that calls Method, e.g. ResolveURL(ctx).
func (r attributeResolver) FormatMethodCall() string {
	return r.Method + "(" + strings.Join(append([]string{"ctx"}, r.Args...), ", ") + ")"
}
