package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
)

const maxBufferSize = 1024 * 1024 * 10 // 10MB

func Parse(fileName string) (TemplateFile, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return TemplateFile{}, err
	}
	fi, err := f.Stat()
	if err != nil {
		return TemplateFile{}, err
	}
	bufferSize := maxBufferSize
	if fi.Size() < int64(bufferSize) {
		bufferSize = int(fi.Size())
	}
	reader := bufio.NewReader(f)
	tfr := NewTemplateFileParser(getDefaultPackageName(fileName)).Parse(input.NewWithBufferSize(reader, bufferSize))
	if tfr.Error != nil {
		return TemplateFile{}, tfr.Error
	}
	return tfr.Item.(TemplateFile), nil
}

func getDefaultPackageName(fileName string) (pkg string) {
	parent := filepath.Base(filepath.Dir(fileName))
	if !isGoIdentifier(parent) {
		return "main"
	}
	return parent
}

func isGoIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if unicode.IsLetter(r) || r == '_' {
			continue
		}
		if i > 0 && unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func ParseString(template string) (TemplateFile, error) {
	tfr := NewTemplateFileParser("main").Parse(input.NewFromString(template))
	if tfr.Error != nil {
		return TemplateFile{}, tfr.Error
	}
	return tfr.Item.(TemplateFile), nil
}

// NewTemplateFileParser creates a new TemplateFileParser.
func NewTemplateFileParser(pkg string) TemplateFileParser {
	return TemplateFileParser{
		DefaultPackage: pkg,
	}
}

var ErrLegacyFileFormat = errors.New("Legacy file format - run templ migrate")

type TemplateFileParser struct {
	DefaultPackage string
}

func (p TemplateFileParser) Parse(pi parse.Input) parse.Result {
	var tf TemplateFile

	// If we're parsing a legacy file, complain that migration needs to happen.
	pr := parse.String("{% package")(pi)
	if pr.Success {
		return parse.Failure("Legacy file format", ErrLegacyFileFormat)
	}

	// Required package.
	// package name
	pr = pkg.Parse(pi)
	if pr.Error != nil {
		return pr
	}
	pkg, ok := pr.Item.(Package)
	if !ok {
		pkg = Package{
			Expression: NewExpression("package "+p.DefaultPackage, NewPosition(), NewPosition()),
		}
	}
	tf.Package = pkg

	// Optional whitespace.
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)

	// Optional templates, CSS, and script templates.
	// templ Name(p Parameter) {
	// css Name() {
	// script Name() {
	// Anything else is assumed to be Go code.
	cssp := newCSSParser()
	stp := newScriptTemplateParser()
	templateContent := parse.Any(template.Parse, cssp.Parse, stp.Parse)
outer:
	for {
		pr := templateContent(pi)
		if pr.Error != nil && pr.Error != io.EOF {
			return pr
		}
		if pr.Success {
			switch pr.Item.(type) {
			case HTMLTemplate:
				tf.Nodes = append(tf.Nodes, pr.Item.(HTMLTemplate))
			case CSSTemplate:
				tf.Nodes = append(tf.Nodes, pr.Item.(CSSTemplate))
			case ScriptTemplate:
				tf.Nodes = append(tf.Nodes, pr.Item.(ScriptTemplate))
			default:
				return parse.Failure("unknown node type", fmt.Errorf("unknown node type %s", reflect.TypeOf(pr.Item).Name()))
			}
			// Eat optional whitespace.
			parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)
			continue
		}
		if pr.Error == io.EOF {
			break
		}

		// Anything that isn't template content is Go code.
		var code strings.Builder
		from := NewPositionFromInput(pi)
		for {
			// Check to see if this line isn't Go code.
			last := NewPositionFromInput(pi)
			l, err := readLine(pi)
			if err != nil && err != io.EOF {
				return parse.Failure("unknown error", err)
			}
			hasTemplatePrefix := strings.HasPrefix(l, "templ ") || strings.HasPrefix(l, "css ") || strings.HasPrefix(l, "script ")
			if hasTemplatePrefix && strings.HasSuffix(l, "{\n") {
				// Unread the line.
				rewind(pi, last.Index)
				// Take the code so far.
				if code.Len() > 0 {
					expr := NewExpression(strings.TrimSpace(code.String()), from, NewPositionFromInput(pi))
					tf.Nodes = append(tf.Nodes, GoExpression{Expression: expr})
				}
				// Carry on parsing.
				break
			}
			code.WriteString(l)
			if err == io.EOF {
				if code.Len() > 0 {
					expr := NewExpression(strings.TrimSpace(code.String()), from, NewPositionFromInput(pi))
					tf.Nodes = append(tf.Nodes, GoExpression{Expression: expr})
				}
				// Stop parsing.
				break outer
			}
		}
	}
	return parse.Success("template file", tf, nil)
}

func readLine(pi parse.Input) (string, error) {
	var sb strings.Builder
	for {
		r, err := pi.Advance()
		if err != nil {
			return sb.String(), err
		}
		sb.WriteRune(r)
		if r == '\n' {
			return sb.String(), err
		}
	}
}
