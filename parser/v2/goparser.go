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
