package generator

import (
	"fmt"
	"html"
	"io"
	"reflect"
	"strings"

	"github.com/a-h/templ"
)

func NewRangeWriter(w io.Writer) *RangeWriter {
	return &RangeWriter{
		Current: templ.NewPosition(),
		w:       w,
	}
}

type RangeWriter struct {
	Current templ.Position
	w       io.Writer
}

func (rw *RangeWriter) Write(s string) (r templ.Range, err error) {
	r.From = templ.Position{
		Index: rw.Current.Index,
		Line:  rw.Current.Line,
		Col:   rw.Current.Col,
	}
	var n int
	for _, c := range s {
		if c == '\n' {
			rw.Current.Line++
			rw.Current.Col = 0
		}
		rw.Current.Col++
		n, err = io.WriteString(rw.w, string(c))
		rw.Current.Index += int64(n)
		if err != nil {
			return r, err
		}
	}
	r.To = rw.Current
	return r, err
}

func Generate(template templ.TemplateFile, w io.Writer) (sm *templ.SourceMap, err error) {
	g := generator{
		tf:        template,
		w:         NewRangeWriter(w),
		sourceMap: templ.NewSourceMap(),
	}
	err = g.generate()
	sm = g.sourceMap
	return
}

type generator struct {
	tf        templ.TemplateFile
	w         *RangeWriter
	sourceMap *templ.SourceMap
}

func (g *generator) generate() error {
	ops := []func() error{
		g.writePackage,
		g.writeImports,
		g.writeTemplates,
	}
	for i := 0; i < len(ops); i++ {
		if err := ops[i](); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writePackage() error {
	var r templ.Range
	var err error
	// package
	if _, err = g.w.Write("package "); err != nil {
		return err
	}
	if r, err = g.w.Write(g.tf.Package.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(g.tf.Package.Expression, r)
	if _, err = g.w.Write("\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeImports() error {
	var r templ.Range
	var err error
	// Always import html because it's used to encode attribute and HTML element content.
	if _, err = g.w.Write("import \"html\"\n"); err != nil {
		return err
	}
	// Always import io because it's the first parameter of a template function.
	if _, err = g.w.Write("import \"io\"\n"); err != nil {
		return err
	}
	for _, im := range g.tf.Imports {
		// import
		if _, err = g.w.Write("import "); err != nil {
			return err
		}
		if r, err = g.w.Write(im.Expression.Value); err != nil {
			return err
		}
		g.sourceMap.Add(im.Expression, r)
		if _, err = g.w.Write("\n"); err != nil {
			return err
		}
	}
	if _, err = g.w.Write("\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplates() error {
	for i := 0; i < len(g.tf.Templates); i++ {
		if err := g.writeTemplate(g.tf.Templates[i]); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeTemplate(t templ.Template) error {
	var r templ.Range
	var err error
	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(t.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Name, r)
	// (w io.Writer,
	if _, err = g.w.Write("(w io.Writer, "); err != nil {
		return err
	}
	// Write parameters.
	if r, err = g.w.Write(t.Parameters.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Parameters, r)
	// ) error {
	if _, err = g.w.Write(") error {\n"); err != nil {
		return err
	}
	// Write out the nodes.
	if err = g.writeNodes(t.Children); err != nil {
		return err
	}
	// return nil
	if _, err = g.w.Write("return nil\n"); err != nil {
		return err
	}
	// }
	if _, err = g.w.Write("}\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeNodes(nodes []templ.Node) error {
	for i := 0; i < len(nodes); i++ {
		if err := g.writeNode(nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeNode(node templ.Node) error {
	switch n := node.(type) {
	case templ.Element:
		g.writeElement(n)
	case templ.Whitespace:
		g.writeWhitespace(n)
	case templ.StringExpression:
		g.writeStringExpression(n)
	case templ.ForExpression:
		g.writeForExpression(n)
	case templ.CallTemplateExpression:
		g.writeCallTemplateExpression(n)
	case templ.IfExpression:
		g.writeIfExpression(n)
	default:
		g.w.Write(fmt.Sprintf("Unhandled type: %v\n", reflect.TypeOf(n)))
	}

	//TODO: SwitchExpression.
	return nil
}

func (g *generator) writeIfExpression(n templ.IfExpression) error {
	var r templ.Range
	var err error
	// if
	if _, err = g.w.Write(`if `); err != nil {
		return err
	}
	// x == y
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// Then.
	// {
	if _, err = g.w.Write(`{ ` + "\n"); err != nil {
		return err
	}
	g.writeNodes(n.Then)
	// } else {
	if _, err = g.w.Write(`} else {` + "\n"); err != nil {
		return err
	}
	g.writeNodes(n.Else)
	// }
	if _, err = g.w.Write(`}` + "\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeCallTemplateExpression(n templ.CallTemplateExpression) error {
	var r templ.Range
	var err error
	// Function name.
	if r, err = g.w.Write(n.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Name, r)
	// (w,
	if _, err = g.w.Write(`(w, `); err != nil {
		return err
	}
	// Arguments expression.
	if r, err = g.w.Write(n.Arguments.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Arguments, r)
	// Close up arguments.
	if _, err = g.w.Write(`)` + "\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeForExpression(n templ.ForExpression) error {
	var r templ.Range
	var err error
	// if
	if _, err = g.w.Write(`for `); err != nil {
		return err
	}
	// i, v := range p.Stuff
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(`{ ` + "\n"); err != nil {
		return err
	}
	// Children.
	g.writeNodes(n.Children)
	// }
	if _, err = g.w.Write(`}` + "\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeElement(n templ.Element) error {
	var r templ.Range
	var err error
	// Attributes.
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, "<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
	} else {
		// <div
		if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, "<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		for i := 0; i < len(n.Attributes); i++ {
			switch attr := n.Attributes[i].(type) {
			case templ.ConstantAttribute:
				name := html.EscapeString(attr.Name)
				value := html.EscapeString(attr.Value)
				if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, " %s=\"%s\"")`+"\n", name, value)); err != nil {
					return err
				}
			case templ.ExpressionAttribute:
				name := html.EscapeString(attr.Name)
				// Name
				if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, " %s=")`+"\n", name)); err != nil {
					return err
				}
				// Value.
				// Open quote.
				if _, err = g.w.Write(`io.WriteString(w, "\"")` + "\n"); err != nil {
					return err
				}
				// io.WriteString(w, html.EscapeString(
				if _, err = g.w.Write("io.WriteString(w, html.EscapeString("); err != nil {
					return err
				}
				// p.Name()
				if r, err = g.w.Write(attr.Value.Expression.Value); err != nil {
					return err
				}
				g.sourceMap.Add(attr.Value.Expression, r)
				// ))
				if _, err = g.w.Write("))\n"); err != nil {
					return err
				}
				// Close quote.
				if _, err = g.w.Write(`io.WriteString(w, "\"")` + "\n"); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown attribute type %s", reflect.TypeOf(n.Attributes[i]))
			}
		}
		// >
		if _, err = g.w.Write(`io.WriteString(w, ">")` + "\n"); err != nil {
			return err
		}
	}
	// Children.
	g.writeNodes(n.Children)
	// </div>
	if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, "</%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeWhitespace(n templ.Whitespace) error {
	var err error
	var spaces strings.Builder
	for _, r := range n.Value {
		switch r {
		case '\n':
			spaces.WriteString(`\n`)
		case '\r':
			spaces.WriteString(`\r`)
		case '\t':
			spaces.WriteString(`\t`)
		default:
			spaces.WriteRune(r)
		}
	}
	// io.WriteString(w, "<spaces>")
	if _, err = g.w.Write(fmt.Sprintf(`io.WriteString(w, "%s")`+"\n", spaces.String())); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStringExpression(n templ.StringExpression) error {
	var r templ.Range
	var err error
	// io.WriteString(w, html.EscapeString(
	if _, err = g.w.Write("io.WriteString(w, html.EscapeString("); err != nil {
		return err
	}
	// p.Name()
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// ))
	if _, err = g.w.Write("))\n"); err != nil {
		return err
	}
	return nil
}
