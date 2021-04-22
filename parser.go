package templ

import (
	"fmt"
	"unicode"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
)

// Constants.
var tagEnd = parse.String(" %}")
var newLine = parse.Or(parse.String("\r\n"), parse.Rune('\n'))

// Whitespace.
func asWhitespace(parts []interface{}) (result interface{}, ok bool) {
	var w Whitespace
	for _, ip := range parts {
		w.Value += string(ip.(rune))
	}
	return w, true
}

var whitespaceParser = parse.AtLeast(asWhitespace, 1, parse.RuneInRanges(unicode.White_Space))

// Template

func newTemplateParser() templateParser {
	return templateParser{}
}

type templateParser struct {
}

func (p templateParser) asTemplate(parts []interface{}) (result interface{}, ok bool) {
	te := parts[0].(templateExpression)
	t := Template{
		Name:       te.Name,
		Parameters: te.Parameters,
	}
	t.Children = parts[1].([]Node)
	return t, true
}

func (p templateParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asTemplate,
		newTemplateExpressionParser().Parse, // {% templ FuncName(p Person, other Other) %}
		newTemplateNodeParser().Parse,       // template whitespace, if/switch/for, or node string expression
		parse.String("{% endtmpl %}"),       // {% endtempl %}
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
	)(pi)
}

// Template node (element, call, if, switch, for, whitespace etc.)
func newTemplateNodeParser() templateNodeParser {
	return templateNodeParser{}
}

type templateNodeParser struct {
}

func (p templateNodeParser) asTemplateNodeArray(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Node, len(parts))
	for i := 0; i < len(parts); i++ {
		op[i] = parts[i].(Node)
	}
	return op, true
}

func (p templateNodeParser) Parse(pi parse.Input) parse.Result {
	return parse.AtLeast(p.asTemplateNodeArray, 0, parse.Any(
		newElementParser().Parse, // <a>, <br/> etc.
		whitespaceParser,
		newStringExpressionParser().Parse,       // {%= strings.ToUpper("abc") %}
		newIfExpressionParser().Parse,           // if {}
		newForExpressionParser().Parse,          // for {}
		newCallTemplateExpressionParser().Parse, // {% call TemplateName(a, b, c) %}
	))(pi)
}

// Parse error.
func newParseError(msg string, from Position, to Position) parseError {
	return parseError{
		Message: msg,
		From:    from,
		To:      to,
	}
}

type parseError struct {
	Message string
	From    Position
	To      Position
}

func (pe parseError) Error() string {
	return fmt.Sprintf("%v from %v to %v", pe.Message, pe.From, pe.To)
}

// TemplateFile.
func NewTemplateFileParser() TemplateFileParser { return TemplateFileParser{} }

type TemplateFileParser struct {
}

func (p TemplateFileParser) asImportArray(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Import{}, true
	}
	return parts[0].([]Import), true
}

func (p TemplateFileParser) asTemplateArray(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Template{}, true
	}
	return parts[0].([]Template), true
}

func (p TemplateFileParser) Parse(pi parse.Input) parse.Result {
	var tf TemplateFile
	from := NewPositionFromInput(pi)

	// Required package.
	// {% package name %}
	pr := newPackageParser().Parse(pi)
	if pr.Error != nil {
		return pr
	}
	if !pr.Success {
		return parse.Failure("TemplateFileParser", newParseError("package not found", from, NewPositionFromInput(pi)))
	}
	tf.Package = pr.Item.(Package)

	// Optional whitespace.
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)

	// Optional imports.
	// {% import "strings" %}
	ipr := parse.Many(p.asImportArray, 0, -1, newImportParser().Parse)(pi)
	if ipr.Error != nil {
		return ipr
	}
	tf.Imports = ipr.Item.([]Import)

	// Optional whitespace.
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)

	// Optional templates.
	// {% templ Name(p Parameter) %}
	tr := parse.Optional(p.asTemplateArray,
		newTemplateParser().Parse,
	)(pi)
	if tr.Error != nil {
		return tr
	}
	tf.Templates = tr.Item.([]Template)

	// Success.
	return parse.Success("template file", tf, nil)
}

func ParseString(template string) (TemplateFile, *SourceMap, error) {
	srl := NewSourceMap()
	tfr := NewTemplateFileParser().Parse(input.NewFromString(template))
	if tfr.Error != nil {
		return TemplateFile{}, srl, tfr.Error
	}
	return tfr.Item.(TemplateFile), srl, nil
}
