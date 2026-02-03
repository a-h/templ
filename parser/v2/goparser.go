package parser

import (
	"fmt"
	"strings"

	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

func parseGoFuncDecl(prefix string, pi *parse.Input) (name string, expression Expression, err error) {
	prefix = prefix + " "
	from := pi.Index()
	src, _ := pi.Peek(-1)
	src = strings.TrimPrefix(src, prefix)
	name, expr, err := goexpression.Func("func " + src)
	if err != nil {
		return name, expression, parse.Error(fmt.Sprintf("invalid %s declaration: %v", prefix, err.Error()), pi.Position())
	}
	pi.Take(len(prefix) + len(expr))
	to := pi.Position()
	return name, NewExpression(expr, pi.PositionAt(from+len(prefix)), to), nil
}

func parseTemplFuncDecl(pi *parse.Input) (name string, expression Expression, err error) {
	return parseGoFuncDecl("templ", pi)
}

func parseCSSFuncDecl(pi *parse.Input) (name string, expression Expression, err error) {
	return parseGoFuncDecl("css", pi)
}

// parseAnonymousTemplFuncParams parses the parameters of an anonymous templ function.
// e.g., "templ(x string)" returns "(x string)"
func parseAnonymousTemplFuncParams(pi *parse.Input) (expression Expression, err error) {
	from := pi.Index()
	src, _ := pi.Peek(-1)
	// "templ(x string)" -> "func(x string)" for parsing
	src = strings.TrimPrefix(src, "templ")
	expr, err := goexpression.AnonymousFuncParams("func" + src)
	if err != nil {
		return expression, parse.Error(fmt.Sprintf("invalid anonymous templ parameters: %v", err.Error()), pi.Position())
	}
	pi.Take(len("templ") + len(expr))
	to := pi.Position()
	return NewExpression(expr, pi.PositionAt(from+len("templ")), to), nil
}

func parseGoSliceArgs(pi *parse.Input) (r Expression, err error) {
	from := pi.Position()
	src, _ := pi.Peek(-1)
	expr, err := goexpression.SliceArgs(src)
	if err != nil {
		return r, err
	}
	pi.Take(len(expr))
	to := pi.Position()
	return NewExpression(expr, from, to), nil
}

func peekPrefix(pi *parse.Input, prefixes ...string) bool {
	for _, prefix := range prefixes {
		pp, ok := pi.Peek(len(prefix))
		if !ok {
			continue
		}
		if prefix == pp {
			return true
		}
	}
	return false
}

type extractor func(content string) (start, end int, err error)

func parseGo(name string, pi *parse.Input, e extractor) (r Expression, err error) {
	from := pi.Index()
	src, _ := pi.Peek(-1)
	start, end, err := e(src)
	if err != nil {
		return r, parse.Error(fmt.Sprintf("%s: invalid go expression: %v", name, err.Error()), pi.Position())
	}
	expr := src[start:end]
	pi.Take(end)
	return NewExpression(expr, pi.PositionAt(from+start), pi.PositionAt(from+end)), nil
}

// EmbeddedTemplateInfo represents an embedded templ() block found in an expression.
type EmbeddedTemplateInfo struct {
	// StartIndex is the index in the expression where the templ( starts.
	StartIndex int
	// EndIndex is the index in the expression after the closing }.
	EndIndex int
	// Params is the parameter string, e.g., "()" or "(name string)".
	Params string
	// Body is the template body content.
	Body string
}

// findEmbeddedTemplates scans an expression string for embedded templ() blocks.
// It returns information about each found block.
func findEmbeddedTemplates(expr string) []EmbeddedTemplateInfo {
	var results []EmbeddedTemplateInfo
	i := 0
	for i < len(expr) {
		// Look for "templ("
		idx := strings.Index(expr[i:], "templ(")
		if idx == -1 {
			break
		}
		startIdx := i + idx

		// Check that this is not part of a longer identifier (e.g., "mytempl(")
		if startIdx > 0 {
			prevChar := expr[startIdx-1]
			if isIdentChar(prevChar) {
				i = startIdx + 1
				continue
			}
		}

		// Parse from templ( to find the matching ) and then {
		rest := expr[startIdx:]
		info, ok := parseEmbeddedTemplate(rest, startIdx)
		if !ok {
			i = startIdx + 1
			continue
		}

		results = append(results, info)
		i = info.EndIndex
	}
	return results
}

// isIdentChar returns true if c is a valid identifier character.
func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// parseEmbeddedTemplate parses a templ() { ... } block from the given string.
// It returns the parsed info and true if successful, or false if not a valid embedded template.
func parseEmbeddedTemplate(s string, baseIndex int) (EmbeddedTemplateInfo, bool) {
	// s starts with "templ("
	if !strings.HasPrefix(s, "templ(") {
		return EmbeddedTemplateInfo{}, false
	}

	// Find the matching ) for the params
	i := len("templ(")
	depth := 1
	for i < len(s) && depth > 0 {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '"':
			// Skip string literal
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					i++
				}
				i++
			}
		case '`':
			// Skip raw string literal
			i++
			for i < len(s) && s[i] != '`' {
				i++
			}
		}
		i++
	}
	if depth != 0 {
		return EmbeddedTemplateInfo{}, false
	}

	paramsEnd := i
	params := s[len("templ"):paramsEnd] // "(params)"

	// Skip whitespace before {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
		i++
	}

	// Expect {
	if i >= len(s) || s[i] != '{' {
		return EmbeddedTemplateInfo{}, false
	}

	// Find the matching }
	braceStart := i
	i++
	depth = 1
	for i < len(s) && depth > 0 {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
		case '"':
			// Skip string literal
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					i++
				}
				i++
			}
		case '`':
			// Skip raw string literal
			i++
			for i < len(s) && s[i] != '`' {
				i++
			}
		}
		i++
	}
	if depth != 0 {
		return EmbeddedTemplateInfo{}, false
	}

	braceEnd := i
	body := s[braceStart+1 : braceEnd-1] // Content inside { }

	return EmbeddedTemplateInfo{
		StartIndex: baseIndex,
		EndIndex:   baseIndex + braceEnd,
		Params:     params,
		Body:       strings.TrimSpace(body),
	}, true
}
