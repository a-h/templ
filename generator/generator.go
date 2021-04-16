package generator

import (
	"fmt"
	"io"

	"github.com/a-h/templ"
)

func Generate(template templ.TemplateFile, w io.Writer) error {
	g := generator{
		tf: template,
		w:  w,
	}
	return g.generate()
}

type generator struct {
	tf templ.TemplateFile
	w  io.Writer
}

func (g *generator) generate() error {
	ops := []func() error{
		g.writePackage,
		g.writeNewLine,
		g.writeImports,
		g.writeNewLine,
		g.writeTemplates,
	}
	for i := 0; i < len(ops); i++ {
		if err := ops[i](); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeStringf(s string, v ...interface{}) error {
	return g.writeString(fmt.Sprintf(s, v...))
}

func (g *generator) writeString(s string) error {
	_, err := io.WriteString(g.w, s)
	return err
}

func (g *generator) writeNewLine() error {
	return g.writeString("\n")
}

func (g *generator) writePackage() error {
	return g.writeStringf("package %s\n", g.tf.Package.Expression)
}

func (g *generator) writeImports() error {
	for _, im := range g.tf.Imports {
		if err := g.writeStringf("import %s\n", im.Expression); err != nil {
			return err
		}
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
	//TODO: Update the target map.
	g.writeString("func %s(%s) {\n")
	return nil
}
