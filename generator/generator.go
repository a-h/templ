package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/parser/v2"
)

func Generate(template parser.TemplateFile, w io.Writer) (sm *parser.SourceMap, err error) {
	g := generator{
		tf:        template,
		w:         NewRangeWriter(w),
		sourceMap: parser.NewSourceMap(),
	}
	err = g.generate()
	sm = g.sourceMap
	return
}

type generator struct {
	tf          parser.TemplateFile
	w           *RangeWriter
	sourceMap   *parser.SourceMap
	variableID  int
	childrenVar string
}

func (g *generator) generate() (err error) {
	if err = g.writeCodeGeneratedComment(); err != nil {
		return
	}
	if err = g.writePackage(); err != nil {
		return
	}
	if err = g.writeImports(); err != nil {
		return
	}
	if err = g.writeTemplateNodes(); err != nil {
		return
	}
	return err
}

// Binary builds set this version string. goreleaser sets the value using Go build ldflags.
var version string

// Source builds use this value. When installed using `go install github.com/a-h/templ/cmd/templ@latest` the `version` variable is empty, but
// the debug.ReadBuildInfo return value provides the package version number installed by `go install`
func goInstallVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

func getVersion() string {
	if version != "" {
		return version
	}
	return goInstallVersion()
}

func (g *generator) writeCodeGeneratedComment() error {
	_, err := g.w.Write(fmt.Sprintf("// Code generated by templ@%s DO NOT EDIT.\n\n", getVersion()))
	return err
}

func (g *generator) writePackage() error {
	var r parser.Range
	var err error
	// package ...
	if r, err = g.w.Write(g.tf.Package.Expression.Value + "\n\n"); err != nil {
		return err
	}
	g.sourceMap.Add(g.tf.Package.Expression, r)
	if _, err = g.w.Write("//lint:file-ignore SA4006 This context is only used if a nested component is present.\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) templateNodeInfo() (hasTemplates bool, hasCSS bool) {
	for _, n := range g.tf.Nodes {
		switch n.(type) {
		case parser.HTMLTemplate:
			hasTemplates = true
		case parser.CSSTemplate:
			hasCSS = true
		}
		if hasTemplates && hasCSS {
			return
		}
	}
	return
}

func (g *generator) writeImports() error {
	var err error
	// Always import templ because it's the interface type of all templates.
	if _, err = g.w.Write("import \"github.com/a-h/templ\"\n"); err != nil {
		return err
	}
	hasTemplates, hasCSS := g.templateNodeInfo()
	if hasTemplates {
		// The first parameter of a template function.
		if _, err = g.w.Write("import \"context\"\n"); err != nil {
			return err
		}
		// The second parameter of a template function.
		if _, err = g.w.Write("import \"io\"\n"); err != nil {
			return err
		}
		// Buffer namespace.
		if _, err = g.w.Write("import \"bytes\"\n"); err != nil {
			return err
		}
	}
	if hasCSS {
		// strings.Builder is used to create CSS.
		if _, err = g.w.Write("import \"strings\"\n"); err != nil {
			return err
		}
	}
	if _, err = g.w.Write("\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplateNodes() error {
	for i := 0; i < len(g.tf.Nodes); i++ {
		switch n := g.tf.Nodes[i].(type) {
		case parser.GoExpression:
			if err := g.writeGoExpression(n); err != nil {
				return err
			}
		case parser.HTMLTemplate:
			if err := g.writeTemplate(n); err != nil {
				return err
			}
		case parser.CSSTemplate:
			if err := g.writeCSS(n); err != nil {
				return err
			}
		case parser.ScriptTemplate:
			if err := g.writeScript(n); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown node type: %v", reflect.TypeOf(n))
		}
	}
	return nil
}

func (g *generator) writeCSS(n parser.CSSTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(n.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Name, r)
	// () templ.CSSClass {
	if _, err = g.w.Write("() templ.CSSClass {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// var templCSSBuilder strings.Builder
		if _, err = g.w.WriteIndent(indentLevel, "var templCSSBuilder strings.Builder\n"); err != nil {
			return err
		}
		for i := 0; i < len(n.Properties); i++ {
			switch p := n.Properties[i].(type) {
			case parser.ConstantCSSProperty:
				// Carry out sanitization at compile time for constants.
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSBuilder.WriteString(`%s`)\n", templ.SanitizeCSS(p.Name, p.Value))); err != nil {
					return err
				}
			case parser.ExpressionCSSProperty:
				// templCSSBuilder.WriteString(templ.SanitizeCSS('name', p.Expression()))
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSBuilder.WriteString(string(templ.SanitizeCSS(`%s`, ", p.Name)); err != nil {
					return err
				}
				if r, err = g.w.Write(p.Value.Expression.Value); err != nil {
					return err
				}
				g.sourceMap.Add(p.Value.Expression, r)
				if _, err = g.w.Write(")))\n"); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown CSS property type: %v", reflect.TypeOf(p))
			}
		}
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSID := templ.CSSID(`%s`, templCSSBuilder.String())\n", n.Name.Value)); err != nil {
			return err
		}
		// return templ.CSS {
		if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentCSSClass{\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// ID: templCSSID,
			if _, err = g.w.WriteIndent(indentLevel, "ID: templCSSID,\n"); err != nil {
				return err
			}
			// Class: templ.SafeCSS(".cssID{" + templ.CSSBuilder.String() + "}"),
			if _, err = g.w.WriteIndent(indentLevel, "Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeGoExpression(n parser.GoExpression) (err error) {
	if _, err = g.w.WriteIndent(0, "// GoExpression\n"); err != nil {
		return err
	}
	r, err := g.w.Write(n.Expression.Value)
	if err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	if _, err = g.w.WriteIndent(0, "\n\n"); err != nil {
		return err
	}
	return err
}

func (g *generator) writeTemplate(t parser.HTMLTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	// (r *Receiver) Name(params []string)
	if r, err = g.w.Write(t.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Expression, r)
	// templ.Component {
	if _, err = g.w.Write(" templ.Component {\n"); err != nil {
		return err
	}
	indentLevel++
	// return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
	if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if _, err = g.w.WriteIndent(indentLevel, "templBuffer, templIsBuffer := w.(*bytes.Buffer)\n"); err != nil {
			return err
		}
		if _, err = g.w.WriteIndent(indentLevel, "if !templIsBuffer {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// templBuffer = templ.GetBuffer()
			if _, err = g.w.WriteIndent(indentLevel, "templBuffer = templ.GetBuffer()\n"); err != nil {
				return err
			}
			// defer templ.ReleaseBuffer(templBuffer)
			if _, err = g.w.WriteIndent(indentLevel, "defer templ.ReleaseBuffer(templBuffer)\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		// ctx = templ.InitializeContext(ctx)
		if _, err = g.w.WriteIndent(indentLevel, "ctx = templ.InitializeContext(ctx)\n"); err != nil {
			return err
		}
		g.childrenVar = g.createVariableName()
		// var_1 := templ.GetChildren(ctx)
		//  if var_1 == nil {
		//  	var_1 = templ.NopComponent
		// }
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("%s := templ.GetChildren(ctx)\n", g.childrenVar)); err != nil {
			return err
		}
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("if %s == nil {\n", g.childrenVar)); err != nil {
			return err
		}
		{
			indentLevel++
			if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("%s = templ.NopComponent\n", g.childrenVar)); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		// ctx = templ.ClearChildren(children)
		if _, err = g.w.WriteIndent(indentLevel, "ctx = templ.ClearChildren(ctx)\n"); err != nil {
			return err
		}
		// Nodes.
		if err = g.writeNodes(indentLevel, nil, stripWhitespace(t.Children)); err != nil {
			return err
		}
		// Return the buffer.
		if _, err = g.w.WriteIndent(indentLevel, "if !templIsBuffer {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// _, err = io.Copy(w, templBuffer)
			if _, err = g.w.WriteIndent(indentLevel, "_, err = io.Copy(w, templBuffer)\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		// return nil
		if _, err = g.w.WriteIndent(indentLevel, "return err\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// })
	if _, err = g.w.WriteIndent(indentLevel, "})\n"); err != nil {
		return err
	}
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func stripNonCriticalElementWhitespace(input []parser.Node) (output []parser.Node) {
	// Remove element, whitespace, element
	// Remove element, whitespace, if etc.
	// Retain text, whitespace, element
	// Retain element, whitespace, text
	for i := range input {
		var prev, curr, next parser.Node
		if i > 0 {
			prev = input[i-1]
		}
		curr = input[i]
		if i < len(input)-1 {
			next = input[i+1]
		}
		_, isWhiteSpace := curr.(parser.Whitespace)
		if !isWhiteSpace {
			output = append(output, curr)
			continue
		}
		if prev == nil {
			// Trim start whitespace.
			continue
		}
		if next == nil {
			// Trim end whitespace.
			continue
		}
		_, prevIsText := prev.(parser.Text)
		_, nextIsText := next.(parser.Text)
		if prevIsText || nextIsText {
			// Allow whitespace that includes text.
			output = append(output, curr)
			continue
		}
	}
	return
}

func stripWhitespace(input []parser.Node) (output []parser.Node) {
	for i, n := range input {
		if _, isWhiteSpace := n.(parser.Whitespace); !isWhiteSpace {
			output = append(output, input[i])
		}
	}
	return output
}

func stripLeadingWhitespace(nodes []parser.Node) []parser.Node {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if _, isWhiteSpace := n.(parser.Whitespace); !isWhiteSpace {
			return nodes[i:]
		}
	}
	return []parser.Node{}
}

func stripTrailingWhitespace(nodes []parser.Node) []parser.Node {
	for i := len(nodes) - 1; i >= 0; i-- {
		n := nodes[i]
		if _, isWhiteSpace := n.(parser.Whitespace); !isWhiteSpace {
			return nodes[0 : i+1]
		}
	}
	return []parser.Node{}
}

func stripLeadingAndTrailingWhitespace(nodes []parser.Node) []parser.Node {
	return stripTrailingWhitespace(stripLeadingWhitespace(nodes))
}

func (g *generator) writeNodes(indentLevel int, parent parser.Node, nodes []parser.Node) error {
	for _, n := range nodes {
		if err := g.writeNode(indentLevel, n); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeNode(indentLevel int, current parser.Node) (err error) {
	switch n := current.(type) {
	case parser.DocType:
		err = g.writeDocType(indentLevel, n)
	case parser.Element:
		err = g.writeElement(indentLevel, n)
	case parser.ChildrenExpression:
		err = g.writeChildrenExpression(indentLevel)
	case parser.RawElement:
		err = g.writeRawElement(indentLevel, n)
	case parser.ForExpression:
		err = g.writeForExpression(indentLevel, n)
	case parser.CallTemplateExpression:
		err = g.writeCallTemplateExpression(indentLevel, n)
	case parser.TemplElementExpression:
		err = g.writeTemplElementExpression(indentLevel, n)
	case parser.IfExpression:
		err = g.writeIfExpression(indentLevel, n)
	case parser.SwitchExpression:
		err = g.writeSwitchExpression(indentLevel, n)
	case parser.StringExpression:
		err = g.writeStringExpression(indentLevel, n.Expression)
	case parser.Whitespace:
		err = g.writeWhitespace(indentLevel, n)
	case parser.Text:
		err = g.writeText(indentLevel, n)
	default:
		_, err = g.w.Write(fmt.Sprintf("Unhandled type: %v\n", reflect.TypeOf(n)))
	}
	return
}

func (g *generator) writeDocType(indentLevel int, n parser.DocType) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// DocType\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("_, err = templBuffer.WriteString(`<!doctype %s>`)\n", n.Value)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeIfExpression(indentLevel int, n parser.IfExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// If\n"); err != nil {
		return err
	}
	var r parser.Range
	// if
	if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
		return err
	}
	// x == y {
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}
	indentLevel++
	if err = g.writeNodes(indentLevel, n, stripLeadingAndTrailingWhitespace(n.Then)); err != nil {
		return err
	}
	indentLevel--
	if len(n.Else) > 0 {
		// } else {
		if _, err = g.w.WriteIndent(indentLevel, `} else {`+"\n"); err != nil {
			return err
		}
		indentLevel++
		if err = g.writeNodes(indentLevel, n, stripLeadingAndTrailingWhitespace(n.Else)); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeSwitchExpression(indentLevel int, n parser.SwitchExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Switch\n"); err != nil {
		return err
	}
	var r parser.Range
	// switch
	if _, err = g.w.WriteIndent(indentLevel, `switch `); err != nil {
		return err
	}
	// val
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}

	if len(n.Cases) > 0 {
		for _, c := range n.Cases {
			// case x:
			// default:
			if r, err = g.w.WriteIndent(indentLevel, c.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(c.Expression, r)
			indentLevel++
			if err = g.writeNodes(indentLevel, n, stripLeadingAndTrailingWhitespace(c.Children)); err != nil {
				return err
			}
			indentLevel--
		}
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeChildrenExpression(indentLevel int) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Children\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("err = %s.Render(ctx, templBuffer)\n", g.childrenVar)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// TemplElement\n"); err != nil {
		return err
	}
	if len(n.Children) == 0 {
		return g.writeSelfClosingTemplElementExpression(indentLevel, n)
	}
	return g.writeBlockTemplElementExpression(indentLevel, n)
}

func (g *generator) writeBlockTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	var r parser.Range
	childrenName := g.createVariableName()
	if _, err = g.w.WriteIndent(indentLevel, childrenName+" := templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {\n"); err != nil {
		return err
	}
	indentLevel++
	if err = g.writeNodes(indentLevel, n, stripLeadingAndTrailingWhitespace(n.Children)); err != nil {
		return err
	}
	// return nil
	if _, err = g.w.WriteIndent(indentLevel, "return err\n"); err != nil {
		return err
	}
	indentLevel--
	if _, err = g.w.WriteIndent(indentLevel, "})\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, `err = `); err != nil {
		return err
	}
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(templ.WithChildren(ctx, children), templBuffer)
	if _, err = g.w.Write(".Render(templ.WithChildren(ctx, " + childrenName + "), templBuffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeSelfClosingTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, `err = `); err != nil {
		return err
	}
	// Template expression.
	var r parser.Range
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(ctx, templBuffer)
	if _, err = g.w.Write(".Render(ctx, templBuffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeCallTemplateExpression(indentLevel int, n parser.CallTemplateExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// CallTemplate\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, `err = `); err != nil {
		return err
	}
	// Template expression.
	var r parser.Range
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(ctx, templBuffer)
	if _, err = g.w.Write(".Render(ctx, templBuffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeForExpression(indentLevel int, n parser.ForExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// For\n"); err != nil {
		return err
	}
	var r parser.Range
	// for
	if _, err = g.w.WriteIndent(indentLevel, `for `); err != nil {
		return err
	}
	// i, v := range p.Stuff
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}
	// Children.
	indentLevel++
	if err = g.writeNodes(indentLevel, n, stripLeadingAndTrailingWhitespace(n.Children)); err != nil {
		return err
	}
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeErrorHandler(indentLevel int) (err error) {
	_, err = g.w.WriteIndent(indentLevel, "if err != nil {\n")
	if err != nil {
		return err
	}
	indentLevel++
	_, err = g.w.WriteIndent(indentLevel, "return err\n")
	if err != nil {
		return err
	}
	indentLevel--
	_, err = g.w.WriteIndent(indentLevel, "}\n")
	if err != nil {
		return err
	}
	return err
}

func (g *generator) writeElement(indentLevel int, n parser.Element) (err error) {
	if n.IsVoidElement() {
		return g.writeVoidElement(indentLevel, n)
	}
	return g.writeStandardElement(indentLevel, n)
}

func (g *generator) writeVoidElement(indentLevel int, n parser.Element) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Element (void)\n"); err != nil {
		return err
	}
	if len(n.Children) > 0 {
		return fmt.Errorf("writeVoidElement: void element %q must not have child elements", n.Name)
	}
	if len(n.Attributes) == 0 {
		// <br>
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		// <hr
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteIndent(indentLevel, `_, err = templBuffer.WriteString(">")`+"\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	}
	return err
}

func (g *generator) writeStandardElement(indentLevel int, n parser.Element) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Element (standard)\n"); err != nil {
		return err
	}
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		// <style type="text/css"></style>
		if err = g.writeElementCSS(indentLevel, n); err != nil {
			return err
		}
		// <script type="text/javascript"></script>
		if err = g.writeElementScript(indentLevel, n); err != nil {
			return err
		}
		// <div
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteIndent(indentLevel, `_, err = templBuffer.WriteString(">")`+"\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	}
	// Children.
	if err = g.writeNodes(indentLevel, n, stripNonCriticalElementWhitespace(n.Children)); err != nil {
		return err
	}
	// </div>
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("</%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) writeElementCSS(indentLevel int, n parser.Element) (err error) {
	var r parser.Range
	for i := 0; i < len(n.Attributes); i++ {
		if attr, ok := n.Attributes[i].(parser.ExpressionAttribute); ok {
			name := html.EscapeString(attr.Name)
			if name != "class" {
				continue
			}
			if _, err = g.w.WriteIndent(indentLevel, "// Element CSS\n"); err != nil {
				return err
			}
			// Create a class name for the style.
			// var templCSSClassess templ.CSSClasses =
			classesName := g.createVariableName()
			if _, err = g.w.WriteIndent(indentLevel, "var "+classesName+" templ.CSSClasses = "); err != nil {
				return err
			}
			// p.Name()
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			if _, err = g.w.Write("\n"); err != nil {
				return err
			}
			// Render the CSS before the element if required.
			// err = templ.RenderCSSItems(ctx, templBuffer, templCSSClassess...)
			if _, err = g.w.WriteIndent(indentLevel, "err = templ.RenderCSSItems(ctx, templBuffer, "+classesName+"...)\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			// Rewrite the ExpressionAttribute to point at the new variable.
			attr.Expression = parser.Expression{
				Value: classesName + ".String()",
			}
			n.Attributes[i] = attr
		}
	}
	return err
}

func (g *generator) writeElementScript(indentLevel int, n parser.Element) (err error) {
	var scriptExpressions []string
	for i := 0; i < len(n.Attributes); i++ {
		if attr, ok := n.Attributes[i].(parser.ExpressionAttribute); ok {
			name := html.EscapeString(attr.Name)
			if strings.HasPrefix(name, "on") {
				scriptExpressions = append(scriptExpressions, attr.Expression.Value)
			}
		}
	}
	if len(scriptExpressions) == 0 {
		return
	}
	if _, err = g.w.WriteIndent(indentLevel, "// Element Script\n"); err != nil {
		return err
	}
	// Render the scripts before the element if required.
	// err = templ.RenderScriptItems(ctx, templBuffer, a, b, c)
	if _, err = g.w.WriteIndent(indentLevel, "err = templ.RenderScriptItems(ctx, templBuffer, "+strings.Join(scriptExpressions, ", ")+")\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) writeBoolConstantAttribute(indentLevel int, attr parser.BoolConstantAttribute) (err error) {
	name := html.EscapeString(attr.Name)
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString(" %s")`+"\n", name)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeConstantAttribute(indentLevel int, attr parser.ConstantAttribute) (err error) {
	name := html.EscapeString(attr.Name)
	value := html.EscapeString(attr.Value)
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString(" %s=\"%s\"")`+"\n", name, value)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeBoolExpressionAttribute(indentLevel int, attr parser.BoolExpressionAttribute) (err error) {
	name := html.EscapeString(attr.Name)
	// if
	if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
		return err
	}
	// x == y
	var r parser.Range
	if r, err = g.w.Write(attr.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(attr.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}
	{
		indentLevel++
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString(" %s")`+"\n", name)); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeExpressionAttribute(indentLevel int, elementName string, attr parser.ExpressionAttribute) (err error) {
	attrName := html.EscapeString(attr.Name)
	// Name
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString(" %s=")`+"\n", attrName)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	// Value.
	// Open quote.
	if _, err = g.w.WriteIndent(indentLevel, `_, err = templBuffer.WriteString("\"")`+"\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	if elementName == "a" && attr.Name == "href" {
		vn := g.createVariableName()
		// var vn templ.SafeURL =
		if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" templ.SafeURL = "); err != nil {
			return err
		}
		// p.Name()
		var r parser.Range
		if r, err = g.w.Write(attr.Expression.Value); err != nil {
			return err
		}
		g.sourceMap.Add(attr.Expression, r)
		if _, err = g.w.Write("\n"); err != nil {
			return err
		}
		if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString(templ.EscapeString(string("+vn+")))\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		if strings.HasPrefix(attr.Name, "on") {
			// It's a JavaScript handler, and requires special handling, because we expect a JavaScript expression.
			vn := g.createVariableName()
			// var vn templ.ComponentScript =
			if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" templ.ComponentScript = "); err != nil {
				return err
			}
			// p.Name()
			var r parser.Range
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			if _, err = g.w.Write("\n"); err != nil {
				return err
			}
			if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString("+vn+".Call)\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		} else {
			// templBuffer.WriteString(templ.EscapeString(
			if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString(templ.EscapeString("); err != nil {
				return err
			}
			// p.Name()
			var r parser.Range
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			// ))
			if _, err = g.w.Write("))\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		}
	}
	// Close quote.
	if _, err = g.w.WriteIndent(indentLevel, `_, err = templBuffer.WriteString("\"")`+"\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeElementAttributes(indentLevel int, name string, attrs []parser.Attribute) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Element Attributes\n"); err != nil {
		return err
	}
	for i := 0; i < len(attrs); i++ {
		switch attr := attrs[i].(type) {
		case parser.BoolConstantAttribute:
			if err = g.writeBoolConstantAttribute(indentLevel, attr); err != nil {
				return err
			}
		case parser.ConstantAttribute:
			if err = g.writeConstantAttribute(indentLevel, attr); err != nil {
				return err
			}
		case parser.BoolExpressionAttribute:
			if err = g.writeBoolExpressionAttribute(indentLevel, attr); err != nil {
				return err
			}
		case parser.ExpressionAttribute:
			if err = g.writeExpressionAttribute(indentLevel, name, attr); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown attribute type %s", reflect.TypeOf(attrs[i]))
		}
	}
	return err
}

func (g *generator) writeRawElement(indentLevel int, n parser.RawElement) (err error) {
	if _, err = g.w.WriteIndent(0, "// RawElement\n"); err != nil {
		return err
	}
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		// <div
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteIndent(indentLevel, `_, err = templBuffer.WriteString(">")`+"\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	}
	// Contents.
	if err = g.writeText(0, parser.Text{Value: n.Contents}); err != nil {
		return err
	}
	// </div>
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = templBuffer.WriteString("</%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) createVariableName() string {
	g.variableID++
	return fmt.Sprintf("var_%d", g.variableID)
}

func (g *generator) writeStringExpression(indentLevel int, e parser.Expression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// StringExpression\n"); err != nil {
		return err
	}
	var r parser.Range
	vn := g.createVariableName()
	// var vn string = sExpr
	if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" string = "); err != nil {
		return err
	}
	// p.Name()
	if r, err = g.w.Write(e.Value + "\n"); err != nil {
		return err
	}
	g.sourceMap.Add(e, r)
	// _, err = templBuffer.WriteString(vn)
	if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString(templ.EscapeString("+vn+"))\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeWhitespace(indentLevel int, n parser.Whitespace) (err error) {
	if len(n.Value) == 0 {
		return
	}
	if _, err = g.w.WriteIndent(indentLevel, "// Whitespace (normalised)\n"); err != nil {
		return err
	}
	// _, err = templBuffer.WriteString(` `)
	if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString(` `)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeText(indentLevel int, n parser.Text) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, "// Text\n"); err != nil {
		return err
	}
	vn := g.createVariableName()
	// vn := sExpr
	if _, err = g.w.WriteIndent(indentLevel, vn+" := "+createGoString(n.Value)+"\n"); err != nil {
		return err
	}
	// _, err = templBuffer.WriteString(vn)
	if _, err = g.w.WriteIndent(indentLevel, "_, err = templBuffer.WriteString("+vn+")\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func createGoString(s string) string {
	var sb strings.Builder
	sb.WriteRune('`')
	sects := strings.Split(s, "`")
	for i := 0; i < len(sects); i++ {
		sb.WriteString(sects[i])
		if len(sects) > i+1 {
			sb.WriteString("` + \"`\" + `")
		}
	}
	sb.WriteRune('`')
	return sb.String()
}

func (g *generator) writeScript(t parser.ScriptTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(t.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Name, r)
	// (
	if _, err = g.w.Write("("); err != nil {
		return err
	}
	// Write parameters.
	if r, err = g.w.Write(t.Parameters.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Parameters, r)
	// ) templ.ComponentScript {
	if _, err = g.w.Write(") templ.ComponentScript {\n"); err != nil {
		return err
	}
	indentLevel++
	// return templ.ComponentScript{
	if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentScript{\n"); err != nil {
		return err
	}
	{
		indentLevel++
		fn := functionName(t.Name.Value, t.Value)
		goFn := createGoString(fn)
		// Name: "scriptName",
		if _, err = g.w.WriteIndent(indentLevel, "Name: "+goFn+",\n"); err != nil {
			return err
		}
		// Function: `function scriptName(a, b, c){` + `constantScriptValue` + `}`,
		prefix := "function " + fn + "(" + stripTypes(t.Parameters.Value) + "){"
		suffix := "}"
		if _, err = g.w.WriteIndent(indentLevel, "Function: "+createGoString(prefix+strings.TrimSpace(t.Value)+suffix)+",\n"); err != nil {
			return err
		}
		// Call: templ.SafeScript(scriptName, a, b, c)
		if _, err = g.w.WriteIndent(indentLevel, "Call: templ.SafeScript("+goFn+", "+stripTypes(t.Parameters.Value)+"),\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
		return err
	}
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func functionName(name string, body string) string {
	h := sha256.New()
	h.Write([]byte(body))
	hp := hex.EncodeToString(h.Sum(nil))[0:4]
	return fmt.Sprintf("__templ_%s_%s", name, hp)
}

func stripTypes(parameters string) string {
	variableNames := []string{}
	params := strings.Split(parameters, ",")
	for i := 0; i < len(params); i++ {
		p := strings.Split(strings.TrimSpace(params[i]), " ")
		variableNames = append(variableNames, strings.TrimSpace(p[0]))
	}
	return strings.Join(variableNames, ", ")
}
