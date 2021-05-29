package templ

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
var optionalWhitespaceParser = parse.AtLeast(asWhitespace, 0, parse.RuneInRanges(unicode.White_Space))

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
		parse.String("{% endtempl %}"),      // {% endtempl %}
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
	)(pi)
}

// Parse error.
func newParseError(msg string, from Position, to Position) ParseError {
	return ParseError{
		Message: msg,
		From:    from,
		To:      to,
	}
}

// ParseError details where the error occurred in the file.
type ParseError struct {
	Message string
	From    Position
	To      Position
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("%v at %v", pe.Message, pe.From)
}

// NewTemplateFileParser creates a new TemplateFileParser.
func NewTemplateFileParser() TemplateFileParser { return TemplateFileParser{} }

type TemplateFileParser struct {
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
	ip := newImportParser()
	for {
		ipr := ip.Parse(pi)
		if ipr.Error != nil {
			return ipr
		}
		if !ipr.Success {
			break
		}
		tf.Imports = append(tf.Imports, ipr.Item.(Import))

		// Eat optional whitespace.
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
	}

	// Optional templates.
	// {% templ Name(p Parameter) %}
	tp := newTemplateParser()
	for {
		tpr := tp.Parse(pi)
		if tpr.Error != nil && tpr.Error != io.EOF {
			return tpr
		}
		if !tpr.Success {
			break
		}
		tf.Templates = append(tf.Templates, tpr.Item.(Template))

		// Eat optional whitespace.
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
	}

	// Success.
	return parse.Success("template file", tf, nil)
}

func Parse(fileName string) (TemplateFile, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return TemplateFile{}, err
	}
	reader := bufio.NewReader(f)
	tfr := NewTemplateFileParser().Parse(input.New(reader))
	if tfr.Error != nil {
		return TemplateFile{}, tfr.Error
	}
	return tfr.Item.(TemplateFile), nil
}

func ParseString(template string) (TemplateFile, error) {
	tfr := NewTemplateFileParser().Parse(input.NewFromString(template))
	if tfr.Error != nil {
		return TemplateFile{}, tfr.Error
	}
	return tfr.Item.(TemplateFile), nil
}
