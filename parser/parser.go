package parser

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

func (p templateParser) Parse(pi parse.Input) parse.Result {
	var r HTMLTemplate

	// {% templ FuncName(p Person, other Other) %}
	tepr := newTemplateExpressionParser().Parse(pi)
	if !tepr.Success {
		return tepr
	}
	te := tepr.Item.(templateExpression)
	r.Name = te.Name
	r.Parameters = te.Parameters

	// Once we're in a template, we should expect some template whitespace, if/switch/for,
	// or node string expressions etc.
	from := NewPositionFromInput(pi)
	tnpr := newTemplateNodeParser(endTemplateParser).Parse(pi)
	if tnpr.Error != nil && tnpr.Error != io.EOF {
		return tnpr
	}
	// If there's no match, there's no template elements.
	if !tnpr.Success {
		return parse.Failure("templateParser", newParseError("templ: expected nodes in templ body, but found none", from, NewPositionFromInput(pi)))
	}
	r.Children = tnpr.Item.([]Node)

	// We must have a final {% endtempl %}, or the close has been forgotten.
	// {% endtempl %}
	if et := endTemplateParser(pi); !et.Success {
		return parse.Failure("templateParser", newParseError("templ: missing end (expected '{% endtempl %}')", from, NewPositionFromInput(pi)))
	}

	// Eat optional whitespace.
	if wpr := parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi); wpr.Error != nil {
		return wpr
	}

	return parse.Success("templ", r, nil)
}

var endTemplateParser = parse.String("{% endtempl %}")

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

	// Optional templates, CSS, and script templates.
	// {% templ Name(p Parameter) %}
	// {% css Name() %}
	// {% script Name() %}
	tp := newTemplateParser()
	cssp := newCSSParser()
	stp := newScriptTemplateParser()
	for {
		// Try for a template.
		tpr := tp.Parse(pi)
		if tpr.Error != nil && tpr.Error != io.EOF {
			return tpr
		}
		if tpr.Success {
			tf.Nodes = append(tf.Nodes, tpr.Item.(HTMLTemplate))
			// Eat optional whitespace.
			parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
			continue
		}
		// Try for css.
		cssr := cssp.Parse(pi)
		if cssr.Error != nil && cssr.Error != io.EOF {
			return cssr
		}
		if cssr.Success {
			tf.Nodes = append(tf.Nodes, cssr.Item.(CSSTemplate))
			// Eat optional whitespace.
			parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
			continue
		}
		// Try for script.
		stpr := stp.Parse(pi)
		if stpr.Error != nil && stpr.Error != io.EOF {
			return stpr
		}
		if stpr.Success {
			tf.Nodes = append(tf.Nodes, stpr.Item.(ScriptTemplate))
			// Eat optional whitespace.
			parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
			continue
		}
		break
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
