package parser

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
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

type TemplateFileParser struct {
	DefaultPackage string
}

func (p TemplateFileParser) Parse(pi parse.Input) parse.Result {
	var tf TemplateFile

	// Required package.
	// {% package name %}
	pr := newPackageParser().Parse(pi)
	if pr.Error != nil {
		return pr
	}
	pkg, ok := pr.Item.(Package)
	if !ok {
		pkg = Package{
			Expression: NewExpression(p.DefaultPackage, NewPosition(), NewPosition()),
		}
	}
	tf.Package = pkg

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
