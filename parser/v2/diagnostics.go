package parser

import (
	"errors"
	"fmt"
)

type (
	nodeDiagnoser         func(Node) ([]Diagnostic, error)
	templateFileDiagnoser func(TemplateFile) ([]Diagnostic, error)
)

// Diagnostic for template file.
type Diagnostic struct {
	Message string
	Range   Range
}

func Diagnose(t TemplateFile) ([]Diagnostic, error) {
	var diags []Diagnostic
	var errs error
	for _, d := range diagnosers {
		diag, err := d(t)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		diags = append(diags, diag...)
	}
	return diags, errs
}

var diagnosers = []templateFileDiagnoser{
	templNotImportedDiagnoser,
	diagnoseNodesDiagnoser(
		useOfLegacyCallSyntaxDiagnoser,
		voidElementWithChildrenDiagnoser,
	),
}

func diagnoseNodesDiagnoser(ds ...nodeDiagnoser) templateFileDiagnoser {
	return func(tf TemplateFile) ([]Diagnostic, error) {
		var diags []Diagnostic
		var errs error
		walkTemplate(tf, func(n Node) bool {
			for _, d := range ds {
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

func templNotImportedDiagnoser(tf TemplateFile) ([]Diagnostic, error) {
	if !tf.ContainsTemplImport() {
		return []Diagnostic{{
			Message: "no \"github.com/a-h/templ\" import found. Run `templ fmt .` to fix all instances.",
		}}, nil
	}
	return nil, nil
}
