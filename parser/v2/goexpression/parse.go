package goexpression

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

var ErrContainerFuncNotFound = errors.New("parser error: templ container function not found")
var ErrExpectedNodeNotFound = errors.New("parser error: expected node not found")

var elseRegex = regexp.MustCompile(`^(else )(\s*){`)

func Else(content string) (start, end int, err error) {
	groups := elseRegex.FindStringSubmatch(content)
	if len(groups) == 0 {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return len("else "), len("else ") + len(groups[2]), nil
}

var elseIfRegex = regexp.MustCompile(`^(else\s+)if\s+`)

func ElseIf(content string) (start, end int, err error) {
	groups := elseIfRegex.FindStringSubmatch(content)
	if len(groups) == 0 {
		return 0, 0, ErrExpectedNodeNotFound
	}
	elsePrefix := groups[1]
	start, end, err = extract(content[len(elsePrefix):], IfExtractor{})
	if err != nil {
		return 0, 0, err
	}
	// Since we trimmed the `else ` prefix, we need to add it back on.
	start += len(elsePrefix)
	end += len(elsePrefix)
	return start, end, nil
}

func Case(content string) (start, end int, err error) {
	if !(strings.HasPrefix(content, "case") || strings.HasPrefix(content, "default")) {
		return 0, 0, ErrExpectedNodeNotFound
	}
	prefix := "switch {\n"
	start, end, err = extract(prefix+content+"\n}", CaseExtractor{})
	if err != nil {
		return 0, 0, err
	}
	// Since we added a `switch {` prefix, we need to remove it.
	start -= len(prefix)
	end -= len(prefix)
	return start, end, nil
}

func If(content string) (start, end int, err error) {
	if !strings.HasPrefix(content, "if") {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return extract(content, IfExtractor{})
}

func For(content string) (start, end int, err error) {
	if !strings.HasPrefix(content, "for") {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return extract(content, ForExtractor{})
}

func Switch(content string) (start, end int, err error) {
	if !strings.HasPrefix(content, "switch") {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return extract(content, SwitchExtractor{})
}

func Expression(content string) (start, end int, err error) {
	start, end, err = extract(content, ExprExtractor{})
	if err != nil {
		return 0, 0, err
	}
	// If the expression ends with `...` then it's a child spread expression.
	if suffix := content[end:]; strings.HasPrefix(suffix, "...") {
		end += len("...")
	}
	return start, end, nil
}

type IfExtractor struct{}

var InGoIfExpression = false

func (e IfExtractor) Code(src string, body []ast.Stmt) (start, end int, err error) {
	stmt, ok := body[0].(*ast.IfStmt)
	if !ok {
		return 0, 0, ErrExpectedNodeNotFound
	}
	start = int(stmt.If) + 2
	if stmt.Init != nil {
		end = int(stmt.Init.End()) - 1
	}
	if stmt.Cond != nil {
		end = int(stmt.Cond.End()) - 1
	}
	return start, end, nil
}

type ForExtractor struct{}

func (e ForExtractor) Code(src string, body []ast.Stmt) (start, end int, err error) {
	stmt := body[0]
	switch stmt := stmt.(type) {
	case *ast.ForStmt:
		start = int(stmt.For) + len("for")
		end = int(stmt.Body.Lbrace) - 1
		return start, end, nil
	case *ast.RangeStmt:
		start = int(stmt.For) + len("for")
		end = int(stmt.Body.Lbrace) - 1
		return start, end, nil
	}
	return 0, 0, ErrExpectedNodeNotFound
}

type SwitchExtractor struct{}

func (e SwitchExtractor) Code(src string, body []ast.Stmt) (start, end int, err error) {
	stmt := body[0]
	switch stmt := stmt.(type) {
	case *ast.SwitchStmt:
		start = int(stmt.Switch) + len("switch")
		end = int(stmt.Body.Lbrace) - 1
		return start, end, nil
	case *ast.TypeSwitchStmt:
		start = int(stmt.Switch) + len("switch")
		end = int(stmt.Body.Lbrace) - 1
		return start, end, nil
	}
	return 0, 0, ErrExpectedNodeNotFound
}

type CaseExtractor struct{}

func (e CaseExtractor) Code(src string, body []ast.Stmt) (start, end int, err error) {
	sw, ok := body[0].(*ast.SwitchStmt)
	if !ok {
		return 0, 0, ErrExpectedNodeNotFound
	}
	stmt, ok := sw.Body.List[0].(*ast.CaseClause)
	if !ok {
		return 0, 0, ErrExpectedNodeNotFound
	}
	if stmt.List == nil {
		// Default case.
		start = int(stmt.Case) + len("default")
		end = int(stmt.Colon)
		return start, end, nil
	}
	// Standard case.
	start = int(stmt.Case) + len("case")
	end = int(stmt.Colon) - 1
	return start, end, nil
}

type ExprExtractor struct{}

func (e ExprExtractor) Code(src string, body []ast.Stmt) (start, end int, err error) {
	stmt, ok := body[0].(*ast.ExprStmt)
	if !ok {
		return 0, 0, ErrExpectedNodeNotFound
	}
	start = int(stmt.Pos()) - 1
	end = int(stmt.End()) - 1
	return start, end, nil
}

// Extract a Go expression from the content.
// The Go expression starts at "start" and ends at "end".
// The reader should skip until "length" to pass over the expression and into the next
// logical block.
type Extractor interface {
	Code(src string, body []ast.Stmt) (start, end int, err error)
}

// Func returns the Go code up to the opening brace of the function body.
func Func(content string) (expr string, err error) {
	prefix := "package main\n"
	src := prefix + content

	node, parseErr := parser.ParseFile(token.NewFileSet(), "", src, parser.AllErrors)
	if node == nil {
		return expr, parseErr
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Find the first function declaration.
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		expr, err = src[fn.Pos():fn.Body.Lbrace-1], nil
		return false
	})

	return expr, err
}

func extract(content string, extractor Extractor) (start, end int, err error) {
	prefix := "package main\nfunc templ_container() {\n"
	src := prefix + content

	node, parseErr := parser.ParseFile(token.NewFileSet(), "", src, parser.AllErrors)
	if node == nil {
		return 0, 0, parseErr
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Find the "templ_container" function.
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name.Name != "templ_container" {
			err = ErrContainerFuncNotFound
			return false
		}
		if fn.Body.List == nil || len(fn.Body.List) == 0 {
			return false
		}
		start, end, err = extractor.Code(src, fn.Body.List)
		start -= len(prefix)
		end -= len(prefix)
		return false
	})
	return start, end, err
}
