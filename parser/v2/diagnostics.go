package parser

import (
	"errors"
	"fmt"
)

type diagnoser func(Node) ([]Diagnostic, error)

// Diagnostic for template file.
type Diagnostic struct {
	Message string
	Range   Range
}

func walkTemplate(t TemplateFile, f func(Node) bool) {
	for _, n := range t.Nodes {
		hn, ok := n.(HTMLTemplate)
		if !ok {
			continue
		}
		walkNodes(hn.Children, f)
	}
}
func walkNodes(t []Node, f func(Node) bool) {
	for _, n := range t {
		if !f(n) {
			continue
		}
		if h, ok := n.(CompositeNode); ok {
			walkNodes(h.ChildNodes(), f)
		}
	}
}

var diagnosers = []diagnoser{
	useOfLegacyCallSyntaxDiagnoser,
	voidElementWithChildrenDiagnoser,
}

func Diagnose(t TemplateFile) ([]Diagnostic, error) {
	var diags []Diagnostic
	var errs error
	walkTemplate(t, func(n Node) bool {
		for _, d := range diagnosers {
			diag, err := d(n)
			if err != nil {
				errs = errors.Join(errs, err)
				return false
			}
			diags = append(diags, diag...)
		}
		return true
	})
	return diags, errs
}

func useOfLegacyCallSyntaxDiagnoser(n Node) ([]Diagnostic, error) {
	if c, ok := n.(CallTemplateExpression); ok {
		return []Diagnostic{{
			Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
			Range:   c.Expression.Range,
		}}, nil
	}
	return nil, nil
}

func voidElementWithChildrenDiagnoser(n Node) (d []Diagnostic, err error) {
	e, ok := n.(Element)
	if !ok {
		return
	}
	if !e.IsVoidElement() {
		return
	}
	if len(e.Children) == 0 {
		return
	}
	return []Diagnostic{{
		Message: fmt.Sprintf("void element <%s> should not have child content", e.Name),
		Range:   e.NameRange,
	}}, nil
}
