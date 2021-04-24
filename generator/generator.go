package generator

import (
	"io"

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
	if _, err = g.w.Write(") {\n"); err != nil {
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
	case templ.StringExpression:
		g.writeStringExpression(n)
	}

	//TODO: Whitespace.
	//TODO: Element.
	//TODO: CallTemplateExpression.
	//TODO: SwitchExpression.
	//TODO: ForExpression.
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
