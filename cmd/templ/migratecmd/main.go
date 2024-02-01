package migratecmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	v1 "github.com/a-h/templ/parser/v1"
	v2 "github.com/a-h/templ/parser/v2"
	"github.com/natefinch/atomic"
)

const workerCount = 4

type Arguments struct {
	FileName string
	Path     string
}

func Run(w io.Writer, args Arguments) (err error) {
	if args.FileName != "" {
		return processSingleFile(w, args.FileName)
	}
	return processPath(w, args.Path)
}

func processSingleFile(w io.Writer, fileName string) error {
	start := time.Now()
	err := migrate(fileName)
	fmt.Fprintf(w, "Migrated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func processPath(w io.Writer, path string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(path, migrate, workerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.Error != nil {
			err = errors.Join(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
			errorCount++
			continue
		}
		successCount++
		fmt.Fprintf(w, "%s complete in %v\n", r.FileName, r.Duration)
	}
	fmt.Fprintf(w, "Migrated code for %d templates with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	return err
}

func migrate(fileName string) (err error) {
	// Check that it's actually a V1 file.
	_, err = v2.Parse(fileName)
	if err == nil {
		return fmt.Errorf("migrate: %s able to parse file as V2, are you sure this needs to be migrated?", fileName)
	}
	if err != v2.ErrLegacyFileFormat {
		return fmt.Errorf("migrate: %s unexpected error: %v", fileName, err)
	}
	// Parse.
	v1Template, err := v1.Parse(fileName)
	if err != nil {
		return fmt.Errorf("migrate: %s v1 parsing error: %w", fileName, err)
	}
	// Convert.
	var v2Template v2.TemplateFile

	// Copy the package and any imports.
	sb := new(strings.Builder)
	sb.WriteString("package " + v1Template.Package.Expression.Value)
	sb.WriteString("\n")
	if len(v1Template.Imports) > 0 {
		sb.WriteString("\n")
		for _, imp := range v1Template.Imports {
			sb.WriteString("import ")
			sb.WriteString(imp.Expression.Value)
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
	v2Template.Package.Expression.Value = sb.String()

	// Work through the nodes.
	v2Template.Nodes, err = migrateV1TemplateFileNodesToV2TemplateFileNodes(v1Template.Nodes)
	if err != nil {
		return fmt.Errorf("%s error migrating elements: %w", fileName, err)
	}

	// Write the updated file.
	w := new(bytes.Buffer)
	err = v2Template.Write(w)
	if err != nil {
		return fmt.Errorf("%s formatting error: %w", fileName, err)
	}
	err = atomic.WriteFile(fileName, w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}

func migrateV1TemplateFileNodesToV2TemplateFileNodes(in []v1.TemplateFileNode) (out []v2.TemplateFileNode, err error) {
	if in == nil {
		return
	}
	out = make([]v2.TemplateFileNode, len(in))
	for i, tfn := range in {
		tfn := tfn
		out[i], err = migrateV1TemplateFileNodeToV2TemplateFileNode(tfn)
		if err != nil {
			return
		}
	}
	return
}

func migrateV1TemplateFileNodeToV2TemplateFileNode(in v1.TemplateFileNode) (out v2.TemplateFileNode, err error) {
	switch n := in.(type) {
	case v1.ScriptTemplate:
		return v2.ScriptTemplate{
			Name: v2.Expression{
				Value: n.Name.Value,
			},
			Parameters: v2.Expression{
				Value: n.Parameters.Value,
			},
			Value: n.Value,
		}, nil
	case v1.CSSTemplate:
		var t v2.CSSTemplate
		t.Name.Value = n.Name.Value
		t.Properties = make([]v2.CSSProperty, len(n.Properties))
		for i, p := range n.Properties {
			t.Properties[i], err = migrateV1CSSPropertyToV2CSSProperty(p)
			if err != nil {
				return
			}
		}
		return t, nil
	case v1.HTMLTemplate:
		var t v2.HTMLTemplate
		t.Expression.Value = fmt.Sprintf("%s(%s)", n.Name.Value, n.Parameters.Value)
		t.Children, err = migrateV1NodesToV2Nodes(n.Children)
		if err != nil {
			return
		}
		return t, nil
	}
	return nil, fmt.Errorf("migrate: unknown template file node type: %s.%s", reflect.TypeOf(in).PkgPath(), reflect.TypeOf(in).Name())
}

func migrateV1CSSPropertyToV2CSSProperty(in v1.CSSProperty) (out v2.CSSProperty, err error) {
	switch p := in.(type) {
	case v1.ConstantCSSProperty:
		return v2.ConstantCSSProperty{Name: p.Name, Value: p.Value}, nil
	case v1.ExpressionCSSProperty:
		var ep v2.ExpressionCSSProperty
		ep.Name = p.Name
		ep.Value.Expression.Value = p.Value.Expression.Value
		return ep, nil
	}
	return nil, fmt.Errorf("migrate: unknown CSS property type: %s", reflect.TypeOf(in).Name())
}

func migrateV1NodesToV2Nodes(in []v1.Node) (out []v2.Node, err error) {
	if in == nil {
		return
	}
	out = make([]v2.Node, len(in))
	for i, n := range in {
		out[i], err = migrateV1NodeToV2Node(n)
		if err != nil {
			return
		}
	}
	return
}

func migrateV1NodeToV2Node(in v1.Node) (out v2.Node, err error) {
	switch n := in.(type) {
	case v1.Whitespace:
		return v2.Whitespace{Value: n.Value}, nil
	case v1.DocType:
		return v2.DocType{Value: n.Value}, nil
	case v1.Text:
		return v2.Text{Value: n.Value}, nil
	case v1.Element:
		return migrateV1ElementToV2Element(n)
	case v1.CallTemplateExpression:
		cte := v2.CallTemplateExpression{
			Expression: v2.Expression{
				Value: n.Expression.Value,
			},
		}
		return cte, nil
	case v1.IfExpression:
		return migrateV1IfExpressionToV2IfExpression(n)
	case v1.SwitchExpression:
		return migrateV1SwitchExpressionToV2SwitchExpression(n)
	case v1.ForExpression:
		return migrateV1ForExpressionToV2ForExpression(n)
	case v1.StringExpression:
		se := v2.StringExpression{
			Expression: v2.Expression{
				Value: n.Expression.Value,
			},
		}
		return se, nil
	}
	return nil, fmt.Errorf("migrate: unknown node type: %s", reflect.TypeOf(in).Name())
}

func migrateV1ForExpressionToV2ForExpression(in v1.ForExpression) (out v2.ForExpression, err error) {
	out.Expression.Value = in.Expression.Value
	out.Children, err = migrateV1NodesToV2Nodes(in.Children)
	if err != nil {
		return
	}
	return
}

func migrateV1SwitchExpressionToV2SwitchExpression(in v1.SwitchExpression) (out v2.SwitchExpression, err error) {
	out.Expression.Value = in.Expression.Value
	out.Cases = make([]v2.CaseExpression, len(in.Cases))
	for i, c := range in.Cases {
		ce := v2.CaseExpression{
			Expression: v2.Expression{
				Value: "case " + c.Expression.Value + ":",
			},
		}
		ce.Children, err = migrateV1NodesToV2Nodes(c.Children)
		if err != nil {
			return
		}
		out.Cases[i] = ce
	}
	if in.Default != nil {
		d := v2.CaseExpression{
			Expression: v2.Expression{
				Value: "default:",
			},
		}
		d.Children, err = migrateV1NodesToV2Nodes(in.Default)
		if err != nil {
			return
		}
		out.Cases = append(out.Cases, d)
	}
	return
}

func migrateV1IfExpressionToV2IfExpression(in v1.IfExpression) (out v2.IfExpression, err error) {
	out.Expression.Value = in.Expression.Value
	out.Then, err = migrateV1NodesToV2Nodes(in.Then)
	if err != nil {
		return
	}
	out.Else, err = migrateV1NodesToV2Nodes(in.Else)
	if err != nil {
		return
	}
	return
}

func migrateV1ElementToV2Element(in v1.Element) (out v2.Element, err error) {
	out.Attributes = make([]v2.Attribute, len(in.Attributes))
	for i, attr := range in.Attributes {
		out.Attributes[i], err = migrateV1AttributeToV2Attribute(attr)
		if err != nil {
			return
		}
	}
	out.Children = make([]v2.Node, len(in.Children))
	for i, child := range in.Children {
		out.Children[i], err = migrateV1NodeToV2Node(child)
		if err != nil {
			return
		}
	}
	out.Name = in.Name
	return out, nil
}

func migrateV1AttributeToV2Attribute(in v1.Attribute) (out v2.Attribute, err error) {
	switch attr := in.(type) {
	case v1.BoolConstantAttribute:
		return v2.BoolConstantAttribute{Name: attr.Name}, nil
	case v1.ConstantAttribute:
		return v2.ConstantAttribute{Name: attr.Name, Value: attr.Value}, nil
	case v1.BoolExpressionAttribute:
		bea := v2.BoolExpressionAttribute{
			Name: attr.Name,
			Expression: v2.Expression{
				Value: attr.Expression.Value,
			},
		}
		return bea, nil
	case v1.ExpressionAttribute:
		ea := v2.ExpressionAttribute{
			Name: attr.Name,
			Expression: v2.Expression{
				Value: attr.Expression.Value,
			},
		}
		return ea, nil
	}
	return nil, fmt.Errorf("migrate: unknown attribute type: %s", reflect.TypeOf(in).Name())
}
