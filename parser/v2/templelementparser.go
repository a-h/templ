package parser

import (
	"fmt"

	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

type templElementExpressionParser struct{}

func (p templElementExpressionParser) Parse(pi *parse.Input) (n Node, matched bool, err error) {
	start := pi.Position()

	// Check the prefix first.
	if _, matched, err = parse.Rune('@').Parse(pi); err != nil || !matched {
		return nil, false, nil
	}

	// Check for anonymous template invocation: @templ(...) { ... }(...)
	if peekPrefix(pi, "templ(") {
		return parseAnonymousTemplateInvocation(pi, start)
	}

	// Parse the Go expression.
	r := &TemplElementExpression{}
	if r.Expression, err = parseGo("templ element", pi, goexpression.TemplExpression); err != nil {
		return r, true, err
	}

	// Check for embedded templ() blocks in the expression and parse them.
	if err = extractEmbeddedTemplates(r); err != nil {
		return r, true, err
	}

	// Once we've got a start expression, check to see if there's an open brace for children. {\n.
	var hasOpenBrace bool
	_, hasOpenBrace, err = openBraceWithOptionalPadding.Parse(pi)
	if err != nil {
		return
	}
	if !hasOpenBrace {
		r.Range = NewRange(start, pi.Position())
		return r, true, nil
	}

	// Once we've had the start of an element's children, we must conclude the block.

	// Node contents.
	np := newTemplateNodeParser(closeBraceWithOptionalPadding, "templ element closing brace")
	var nodes Nodes
	if nodes, matched, err = np.Parse(pi); err != nil || !matched {
		// Populate the nodes anyway, so that the LSP can use them.
		r.Children = nodes.Nodes
		err = parse.Error("@"+r.Expression.Value+": expected nodes, but none were found", pi.Position())
		return r, true, err
	}
	r.Children = nodes.Nodes

	// Read the required closing brace.
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		err = parse.Error("@"+r.Expression.Value+": missing end (expected '}')", pi.Position())
		return r, true, err
	}

	r.Range = NewRange(start, pi.Position())

	return r, true, nil
}

// extractEmbeddedTemplates finds embedded templ() blocks in the expression,
// parses them, and replaces them with placeholder variable names.
func extractEmbeddedTemplates(r *TemplElementExpression) error {
	embeddedInfos := findEmbeddedTemplates(r.Expression.Value)
	if len(embeddedInfos) == 0 {
		return nil
	}

	r.EmbeddedTemplates = make(map[string]*AnonymousTemplate)

	// Process embedded templates in reverse order so indices remain valid
	newExpr := r.Expression.Value
	for i := len(embeddedInfos) - 1; i >= 0; i-- {
		info := embeddedInfos[i]

		// Parse the template body
		bodyInput := parse.NewInput(info.Body)
		np := newTemplateNodeParser(parse.EOF[string](), "embedded template body")
		nodes, _, err := np.Parse(bodyInput)
		if err != nil {
			return fmt.Errorf("failed to parse embedded template body: %w", err)
		}

		// Create a placeholder variable name
		placeholder := fmt.Sprintf("templ_embedded_%d", i)

		// Create the AnonymousTemplate
		anon := &AnonymousTemplate{
			Expression: Expression{
				Value: info.Params,
				Range: r.Expression.Range, // Approximate position
			},
			Children: nodes.Nodes,
		}

		r.EmbeddedTemplates[placeholder] = anon

		// Replace the embedded template with the placeholder
		before := newExpr[:info.StartIndex]
		after := newExpr[info.EndIndex:]
		newExpr = before + placeholder + after
	}

	r.Expression.Value = newExpr
	return nil
}

// parseAnonymousTemplateInvocation parses @templ(...) { ... }(...)
// The "@" has already been consumed. Input starts at "templ(".
func parseAnonymousTemplateInvocation(pi *parse.Input, start parse.Position) (n Node, matched bool, err error) {
	// Parse the anonymous template (templ(...) { ... })
	anonNode, anonMatched, err := anonymousTemplate.Parse(pi)
	if err != nil {
		return anonNode, true, err
	}
	if !anonMatched {
		return nil, false, parse.Error("@templ: expected anonymous template", pi.Position())
	}
	anon := anonNode.(*AnonymousTemplate)

	// Parse the invocation expression (args)
	var callExpr Expression
	if callExpr, err = parseGo("anonymous template invocation", pi, goexpression.TemplExpression); err != nil {
		// If we can't parse a call expression, we still have a valid anonymous template
		// but it's not being invoked - this might be an error depending on context
		return anon, true, err
	}

	r := &AnonymousTemplateInvocation{
		Template:       anon,
		CallExpression: callExpr,
		Range:          NewRange(start, pi.Position()),
	}

	return r, true, nil
}

var templElementExpression templElementExpressionParser
