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
			if err := g.writeTemplate(i, n); err != nil {
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
		// var templ_7745c5c3_CSSBuilder strings.Builder
		if _, err = g.w.WriteIndent(indentLevel, "var templ_7745c5c3_CSSBuilder strings.Builder\n"); err != nil {
			return err
		}
		for i := 0; i < len(n.Properties); i++ {
			switch p := n.Properties[i].(type) {
			case parser.ConstantCSSProperty:
				// Constant CSS property values are not sanitized.
				if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_CSSBuilder.WriteString("+createGoString(p.String(true))+")\n"); err != nil {
					return err
				}
			case parser.ExpressionCSSProperty:
				// templ_7745c5c3_CSSBuilder.WriteString(templ.SanitizeCSS('name', p.Expression()))
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templ_7745c5c3_CSSBuilder.WriteString(string(templ.SanitizeCSS(`%s`, ", p.Name)); err != nil {
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
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templ_7745c5c3_CSSID := templ.CSSID(`%s`, templ_7745c5c3_CSSBuilder.String())\n", n.Name.Value)); err != nil {
			return err
		}
		// return templ.CSS {
		if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentCSSClass{\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// ID: templ_7745c5c3_CSSID,
			if _, err = g.w.WriteIndent(indentLevel, "ID: templ_7745c5c3_CSSID,\n"); err != nil {
				return err
			}
			// Class: templ.SafeCSS(".cssID{" + templ.CSSBuilder.String() + "}"),
			if _, err = g.w.WriteIndent(indentLevel, "Class: templ.SafeCSS(`.` + templ_7745c5c3_CSSID + `{` + templ_7745c5c3_CSSBuilder.String() + `}`),\n"); err != nil {
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

func (g *generator) writeTemplBuffer(indentLevel int) (err error) {
	// templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := w.(*bytes.Buffer)
	if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, "if !templ_7745c5c3_IsBuffer {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// templ_7745c5c3_Buffer = templ.GetBuffer()
		if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Buffer = templ.GetBuffer()\n"); err != nil {
			return err
		}
		// defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		if _, err = g.w.WriteIndent(indentLevel, "defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)\n"); err != nil {
			return err
		}
		indentLevel--
	}
	if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
		return err
	}
	return
}

func (g *generator) writeTemplate(nodeIdx int, t parser.HTMLTemplate) error {
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
	if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentFunc(func(templ_7745c5c3_Ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		if err := g.writeTemplBuffer(indentLevel); err != nil {
			return err
		}
		// ctx = templ.InitializeContext(ctx)
		if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Ctx = templ.InitializeContext(templ_7745c5c3_Ctx)\n"); err != nil {
			return err
		}
		g.childrenVar = g.createVariableName()
		// templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		// if templ_7745c5c3_Var1 == nil {
		//  	templ_7745c5c3_Var1 = templ.NopComponent
		// }
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("%s := templ.GetChildren(templ_7745c5c3_Ctx)\n", g.childrenVar)); err != nil {
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
		if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Ctx = templ.ClearChildren(templ_7745c5c3_Ctx)\n"); err != nil {
			return err
		}
		// Nodes.
		if err = g.writeNodes(indentLevel, stripWhitespace(t.Children)); err != nil {
			return err
		}
		// Return the buffer.
		if _, err = g.w.WriteIndent(indentLevel, "if !templ_7745c5c3_IsBuffer {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// _, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
			if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		// return templ_7745c5c3_Err
		if _, err = g.w.WriteIndent(indentLevel, "return templ_7745c5c3_Err\n"); err != nil {
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

	// Note: gofmt wants to remove a single empty line at the end of a file
	// so we have to make sure we don't output one if this is the last node.
	closingBrace := "}\n\n"
	if nodeIdx+1 >= len(g.tf.Nodes) {
		closingBrace = "}\n"
	}

	if _, err = g.w.WriteIndent(indentLevel, closingBrace); err != nil {
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
		_, prevIsStringExpr := prev.(parser.StringExpression)
		_, nextIsStringExpr := next.(parser.StringExpression)
		if prevIsStringExpr || nextIsStringExpr {
			// Allow whitespace that comes before or after a template expression.
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

func (g *generator) writeNodes(indentLevel int, nodes []parser.Node) error {
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
	case parser.Comment:
		err = g.writeComment(indentLevel, n)
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
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf("<!doctype %s>", n.Value)); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeIfExpression(indentLevel int, n parser.IfExpression) (err error) {
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
	{
		indentLevel++
		if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(n.Then)); err != nil {
			return err
		}
		indentLevel--
	}
	for _, elseIf := range n.ElseIfs {
		// } else if {
		if _, err = g.w.WriteIndent(indentLevel, `} else if `); err != nil {
			return err
		}
		// x == y {
		if r, err = g.w.Write(elseIf.Expression.Value); err != nil {
			return err
		}
		g.sourceMap.Add(elseIf.Expression, r)
		// {
		if _, err = g.w.Write(` {` + "\n"); err != nil {
			return err
		}
		{
			indentLevel++
			if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(elseIf.Then)); err != nil {
				return err
			}
			indentLevel--
		}
	}
	if len(n.Else) > 0 {
		// } else {
		if _, err = g.w.WriteIndent(indentLevel, `} else {`+"\n"); err != nil {
			return err
		}
		{
			indentLevel++
			if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(n.Else)); err != nil {
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

func (g *generator) writeSwitchExpression(indentLevel int, n parser.SwitchExpression) (err error) {
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
			if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(c.Children)); err != nil {
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
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templ_7745c5c3_Err = %s.Render(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer)\n", g.childrenVar)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	if len(n.Children) == 0 {
		return g.writeSelfClosingTemplElementExpression(indentLevel, n)
	}
	return g.writeBlockTemplElementExpression(indentLevel, n)
}

func (g *generator) writeBlockTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	var r parser.Range
	childrenName := g.createVariableName()
	if _, err = g.w.WriteIndent(indentLevel, childrenName+" := templ.ComponentFunc(func(templ_7745c5c3_Ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {\n"); err != nil {
		return err
	}
	indentLevel++
	if err := g.writeTemplBuffer(indentLevel); err != nil {
		return err
	}
	if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(n.Children)); err != nil {
		return err
	}
	// Return the buffer.
	if _, err = g.w.WriteIndent(indentLevel, "if !templ_7745c5c3_IsBuffer {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// _, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
		if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)\n"); err != nil {
			return err
		}
		indentLevel--
	}
	if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
		return err
	}
	// return nil
	if _, err = g.w.WriteIndent(indentLevel, "return templ_7745c5c3_Err\n"); err != nil {
		return err
	}
	indentLevel--
	if _, err = g.w.WriteIndent(indentLevel, "})\n"); err != nil {
		return err
	}
	if _, err = g.w.WriteIndent(indentLevel, `templ_7745c5c3_Err = `); err != nil {
		return err
	}
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(templ.WithChildren(templ_7745c5c3_Ctx, children), templ_7745c5c3_Buffer)
	if _, err = g.w.Write(".Render(templ.WithChildren(templ_7745c5c3_Ctx, " + childrenName + "), templ_7745c5c3_Buffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeSelfClosingTemplElementExpression(indentLevel int, n parser.TemplElementExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, `templ_7745c5c3_Err = `); err != nil {
		return err
	}
	// Template expression.
	var r parser.Range
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer)
	if _, err = g.w.Write(".Render(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeCallTemplateExpression(indentLevel int, n parser.CallTemplateExpression) (err error) {
	if _, err = g.w.WriteIndent(indentLevel, `templ_7745c5c3_Err = `); err != nil {
		return err
	}
	// Template expression.
	var r parser.Range
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer)
	if _, err = g.w.Write(".Render(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeForExpression(indentLevel int, n parser.ForExpression) (err error) {
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
	if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(n.Children)); err != nil {
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
	_, err = g.w.WriteIndent(indentLevel, "if templ_7745c5c3_Err != nil {\n")
	if err != nil {
		return err
	}
	indentLevel++
	_, err = g.w.WriteIndent(indentLevel, "return templ_7745c5c3_Err\n")
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
	if len(n.Children) > 0 {
		return fmt.Errorf("writeVoidElement: void element %q must not have child elements", n.Name)
	}
	if len(n.Attributes) == 0 {
		// <br>
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s>`, html.EscapeString(n.Name))); err != nil {
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
		// <hr
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s`, html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteStringLiteral(indentLevel, `>`); err != nil {
			return err
		}
	}
	return err
}

func (g *generator) writeStandardElement(indentLevel int, n parser.Element) (err error) {
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s>`, html.EscapeString(n.Name))); err != nil {
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
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s`, html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteStringLiteral(indentLevel, `>`); err != nil {
			return err
		}
	}
	// Children.
	if err = g.writeNodes(indentLevel, stripNonCriticalElementWhitespace(n.Children)); err != nil {
		return err
	}
	// </div>
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`</%s>`, html.EscapeString(n.Name))); err != nil {
		return err
	}
	return err
}

func (g *generator) writeAttributeCSS(indentLevel int, attr parser.ExpressionAttribute) (result parser.ExpressionAttribute, ok bool, err error) {
	var r parser.Range
	name := html.EscapeString(attr.Name)
	if name != "class" {
		ok = false
		return
	}
	// Create a class name for the style.
	// The expression can either be expecting a templ.Classes call, or an expression that returns
	// var templ_7745c5c3_CSSClassess = []any{
	classesName := g.createVariableName()
	if _, err = g.w.WriteIndent(indentLevel, "var "+classesName+" = []any{"); err != nil {
		return
	}
	// p.Name()
	if r, err = g.w.Write(attr.Expression.Value); err != nil {
		return
	}
	g.sourceMap.Add(attr.Expression, r)
	// }\n
	if _, err = g.w.Write("}\n"); err != nil {
		return
	}
	// Render the CSS before the element if required.
	// templ_7745c5c3_Err = templ.RenderCSSItems(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer, templ_7745c5c3_CSSClassess...)
	if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Err = templ.RenderCSSItems(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer, "+classesName+"...)\n"); err != nil {
		return
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return
	}
	// Rewrite the ExpressionAttribute to point at the new variable.
	attr.Expression = parser.Expression{
		Value: "templ.CSSClasses(" + classesName + ").String()",
	}
	return attr, true, nil
}

func (g *generator) writeAttributesCSS(indentLevel int, attrs []parser.Attribute) (err error) {
	for i := 0; i < len(attrs); i++ {
		if attr, ok := attrs[i].(parser.ExpressionAttribute); ok {
			attr, ok, err = g.writeAttributeCSS(indentLevel, attr)
			if err != nil {
				return err
			}
			if ok {
				attrs[i] = attr
			}
		}
		if cattr, ok := attrs[i].(parser.ConditionalAttribute); ok {
			err = g.writeAttributesCSS(indentLevel, cattr.Then)
			if err != nil {
				return err
			}
			err = g.writeAttributesCSS(indentLevel, cattr.Else)
			if err != nil {
				return err
			}
			attrs[i] = cattr
		}
	}
	return nil
}

func (g *generator) writeElementCSS(indentLevel int, n parser.Element) (err error) {
	return g.writeAttributesCSS(indentLevel, n.Attributes)
}

func isScriptAttribute(name string) bool {
	for _, prefix := range []string{"on", "hx-on:"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (g *generator) writeElementScript(indentLevel int, n parser.Element) (err error) {
	var scriptExpressions []string
	for i := 0; i < len(n.Attributes); i++ {
		if attr, ok := n.Attributes[i].(parser.ExpressionAttribute); ok {
			name := html.EscapeString(attr.Name)
			if isScriptAttribute(name) {
				scriptExpressions = append(scriptExpressions, attr.Expression.Value)
			}
		}
	}
	if len(scriptExpressions) == 0 {
		return
	}
	// Render the scripts before the element if required.
	// templ_7745c5c3_Err = templ.RenderScriptItems(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer, a, b, c)
	if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_Err = templ.RenderScriptItems(templ_7745c5c3_Ctx, templ_7745c5c3_Buffer, "+strings.Join(scriptExpressions, ", ")+")\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) writeBoolConstantAttribute(indentLevel int, attr parser.BoolConstantAttribute) (err error) {
	name := html.EscapeString(attr.Name)
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(` %s`, name)); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeConstantAttribute(indentLevel int, attr parser.ConstantAttribute) (err error) {
	name := html.EscapeString(attr.Name)
	value := html.EscapeString(attr.Value)
	value = strings.ReplaceAll(value, "\n", "\\n")
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(` %s=\"%s\"`, name, value)); err != nil {
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
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(` %s`, name)); err != nil {
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
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(` %s=`, attrName)); err != nil {
		return err
	}
	// Value.
	// Open quote.
	if _, err = g.w.WriteStringLiteral(indentLevel, `\"`); err != nil {
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
		if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string("+vn+")))\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		if isScriptAttribute(attr.Name) {
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
			if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("+vn+".Call)\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		} else {
			// templ_7745c5c3_Buffer.WriteString(templ.EscapeString(
			if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString("); err != nil {
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
	if _, err = g.w.WriteStringLiteral(indentLevel, `\"`); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeConditionalAttribute(indentLevel int, elementName string, attr parser.ConditionalAttribute) (err error) {
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
		if err = g.writeElementAttributes(indentLevel, elementName, attr.Then); err != nil {
			return err
		}
		indentLevel--
	}
	if len(attr.Else) > 0 {
		// } else {
		if _, err = g.w.WriteIndent(indentLevel, `} else {`+"\n"); err != nil {
			return err
		}
		{
			indentLevel++
			if err = g.writeElementAttributes(indentLevel, elementName, attr.Else); err != nil {
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

func (g *generator) writeElementAttributes(indentLevel int, name string, attrs []parser.Attribute) (err error) {
	for i := 0; i < len(attrs); i++ {
		switch attr := attrs[i].(type) {
		case parser.BoolConstantAttribute:
			err = g.writeBoolConstantAttribute(indentLevel, attr)
		case parser.ConstantAttribute:
			err = g.writeConstantAttribute(indentLevel, attr)
		case parser.BoolExpressionAttribute:
			err = g.writeBoolExpressionAttribute(indentLevel, attr)
		case parser.ExpressionAttribute:
			err = g.writeExpressionAttribute(indentLevel, name, attr)
		case parser.ConditionalAttribute:
			err = g.writeConditionalAttribute(indentLevel, name, attr)
		default:
			err = fmt.Errorf("unknown attribute type %s", reflect.TypeOf(attrs[i]))
		}
	}
	return
}

func (g *generator) writeRawElement(indentLevel int, n parser.RawElement) (err error) {
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s>`, html.EscapeString(n.Name))); err != nil {
			return err
		}
	} else {
		// <div
		if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`<%s`, html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n.Name, n.Attributes); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteStringLiteral(indentLevel, `>`); err != nil {
			return err
		}
	}
	// Contents.
	if err = g.writeText(indentLevel, parser.Text{Value: n.Contents}); err != nil {
		return err
	}
	// </div>
	if _, err = g.w.WriteStringLiteral(indentLevel, fmt.Sprintf(`</%s>`, html.EscapeString(n.Name))); err != nil {
		return err
	}
	return err
}

func (g *generator) writeComment(indentLevel int, c parser.Comment) (err error) {
	// <!--
	if _, err = g.w.WriteStringLiteral(indentLevel, "<!--"); err != nil {
		return err
	}
	// Contents.
	if err = g.writeText(indentLevel, parser.Text{Value: c.Contents}); err != nil {
		return err
	}
	// -->
	if _, err = g.w.WriteStringLiteral(indentLevel, "-->"); err != nil {
		return err
	}
	return err
}

func (g *generator) createVariableName() string {
	g.variableID++
	return fmt.Sprintf("templ_7745c5c3_Var%d", g.variableID)
}

func (g *generator) writeStringExpression(indentLevel int, e parser.Expression) (err error) {
	if strings.TrimSpace(e.Value) == "" {
		return
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
	// _, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(vn)
	if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString("+vn+"))\n"); err != nil {
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
	// _, err = templ_7745c5c3_Buffer.WriteString(` `)
	if _, err = g.w.WriteStringLiteral(indentLevel, " "); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeText(indentLevel int, n parser.Text) (err error) {
	vn := g.createVariableName()
	// vn := sExpr
	if _, err = g.w.WriteIndent(indentLevel, vn+" := "+createGoString(n.Value)+"\n"); err != nil {
		return err
	}
	// _, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(vn)
	if _, err = g.w.WriteIndent(indentLevel, "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("+vn+")\n"); err != nil {
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
