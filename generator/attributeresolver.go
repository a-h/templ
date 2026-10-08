package generator

import "strconv"

// attributeResolver is the generated code that resolves the value of an
// attribute expression: Constructor wraps the result of the expression, and
// Method converts it to an HTML-escaped string, e.g.
// templruntime.NewAttributeExpression(p.URL).ResolveURL(ctx).
type attributeResolver struct {
	Constructor string
	Method      string
}

var (
	attributeResolverDefault = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "Resolve(ctx)",
	}
	attributeResolverURL = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveURL(ctx)",
	}
	attributeResolverAnimationValue = attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveAnimationValue(ctx)",
	}
	attributeResolverSrcdoc = attributeResolver{
		Constructor: "templruntime.NewSrcdocExpression",
		Method:      "Resolve(ctx)",
	}
)

func newDynamicAttributeResolver(elementName, nameVariable string) attributeResolver {
	return attributeResolver{
		Constructor: "templruntime.NewAttributeExpression",
		Method:      "ResolveDynamic(ctx, " + strconv.Quote(elementName) + ", " + nameVariable + ")",
	}
}
