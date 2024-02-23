package goexpression

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

var ErrContainerFuncNotFound = errors.New("parser error: templ container function not found")
var ErrExpectedNodeNotFound = errors.New("parser error: expected node not found")

func Case(content string) (start, end int, err error) {
	if !(strings.HasPrefix(content, "case") || strings.HasPrefix(content, "default")) {
		return 0, 0, ErrExpectedNodeNotFound
	}
	prefix := "switch {\n"
	start, end, err = extract(prefix+content+"\n}", func(src string, body []ast.Stmt) (start, end int, err error) {
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
			start = int(stmt.Case) - 1
			end = int(stmt.Colon)
			return start, end, nil
		}
		// Standard case.
		start = int(stmt.Case) - 1
		end = int(stmt.Colon)
		return start, end, nil
	})
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
	return extract(content, func(src string, body []ast.Stmt) (start, end int, err error) {
		stmt, ok := body[0].(*ast.IfStmt)
		if !ok {
			return 0, 0, ErrExpectedNodeNotFound
		}
		start = int(stmt.If) + 2
		end = latestEnd(start, stmt.Init, stmt.Cond)
		return start, end, nil
	})
}

func For(content string) (start, end int, err error) {
	if !strings.HasPrefix(content, "for") {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return extract(content, func(src string, body []ast.Stmt) (start, end int, err error) {
		stmt := body[0]
		switch stmt := stmt.(type) {
		case *ast.ForStmt:
			start = int(stmt.For) + len("for")
			end = latestEnd(start, stmt.Init, stmt.Cond, stmt.Post)
			return start, end, nil
		case *ast.RangeStmt:
			start = int(stmt.For) + len("for")
			end = latestEnd(start, stmt.Key, stmt.Value, stmt.X)
			return start, end, nil
		}
		return 0, 0, ErrExpectedNodeNotFound
	})
}

func Switch(content string) (start, end int, err error) {
	if !strings.HasPrefix(content, "switch") {
		return 0, 0, ErrExpectedNodeNotFound
	}
	return extract(content, func(src string, body []ast.Stmt) (start, end int, err error) {
		stmt := body[0]
		switch stmt := stmt.(type) {
		case *ast.SwitchStmt:
			start = int(stmt.Switch) + len("switch")
			end = latestEnd(start, stmt.Init, stmt.Tag)
			return start, end, nil
		case *ast.TypeSwitchStmt:
			start = int(stmt.Switch) + len("switch")
			end = latestEnd(start, stmt.Init, stmt.Assign)
			return start, end, nil
		}
		return 0, 0, ErrExpectedNodeNotFound
	})
}

func Expression(content string) (start, end int, err error) {
	start, end, err = extract(content, func(src string, body []ast.Stmt) (start, end int, err error) {
		stmt, ok := body[0].(*ast.ExprStmt)
		if !ok {
			return 0, 0, ErrExpectedNodeNotFound
		}
		start = int(stmt.Pos()) - 1
		end = int(stmt.End()) - 1
		return start, end, nil
	})
	if err != nil {
		return 0, 0, err
	}
	// If the expression ends with `...` then it's a child spread expression.
	if suffix := content[end:]; strings.HasPrefix(suffix, "...") {
		end += len("...")
	}
	return start, end, nil
}

func SliceArgs(content string) (expr string, err error) {
	prefix := "package main\nvar templ_args = []any{"
	src := prefix + content + "}"

	node, parseErr := parser.ParseFile(token.NewFileSet(), "", src, parser.AllErrors)
	if node == nil {
		return expr, parseErr
	}

	var from, to int
	var stop bool
	ast.Inspect(node, func(n ast.Node) bool {
		if stop {
			return false
		}
		decl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		stop = true
		from = int(decl.Lbrace)
		to = int(decl.Rbrace) - 1
		for _, e := range decl.Elts {
			to = int(e.End()) - 1
		}
		if to > int(decl.Rbrace)-1 {
			to = int(decl.Rbrace) - 1
		}
		betweenEndAndBrace := src[to : decl.Rbrace-1]
		var hasCodeBetweenEndAndBrace bool
		for _, r := range betweenEndAndBrace {
			if !unicode.IsSpace(r) {
				hasCodeBetweenEndAndBrace = true
				break
			}
		}
		if hasCodeBetweenEndAndBrace {
			to = int(decl.Rbrace) - 1
		}
		return false
	})

	return src[from:to], err
}

// Func returns the Go code up to the opening brace of the function body.
func Func(content string) (name, expr string, err error) {
	prefix := "package main\n"
	src := prefix + content

	node, parseErr := parser.ParseFile(token.NewFileSet(), "", src, parser.AllErrors)
	if node == nil {
		return name, expr, parseErr
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Find the first function declaration.
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		start := int(fn.Pos()) + len("func")
		end := fn.Type.Params.End() - 1
		expr = src[start:end]
		name = fn.Name.Name
		return false
	})

	return name, expr, err
}

func latestEnd(start int, nodes ...ast.Node) (end int) {
	end = start
	for _, n := range nodes {
		if n == nil {
			continue
		}
		if int(n.End()) > end {
			end = int(n.End()) - 1
		}
	}
	return end
}

// Extract a Go expression from the content.
// The Go expression starts at "start" and ends at "end".
// The reader should skip until "length" to pass over the expression and into the next
// logical block.
type Extractor func(src string, body []ast.Stmt) (start, end int, err error)

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
		start, end, err = extractor(src, fn.Body.List)
		start -= len(prefix)
		end -= len(prefix)
		return false
	})
	return start, end, err
}
