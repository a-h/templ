package fuzzing

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/a-h/parse"
	"github.com/a-h/templ/internal/jstokenizer"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/file"
	"github.com/dop251/goja/parser"
	"github.com/google/go-cmp/cmp"
)

// marker stands in for a {{ }} Go expression. It is an identifier in code, and text
// in a string, template literal, regular expression, or comment, so that a source
// that contains it can be parsed by a JavaScript parser.
const marker = "__m__"

// commentMarker replaces a marker to check whether it is in a comment. It has the same
// length as marker, so offsets don't change, and it is a syntax error in code.
const commentMarker = "@@@@@"

var seeds = []string{
	`var a = __m__;`,
	`var a = "__m__";`,
	`var a = '__m__';`,
	`var a = "\"__m__";`,
	"var a = `__m__`;",
	"var a = `${ __m__ }`;",
	"var a = `${ b }__m__`;",
	"var a = `${ `${ __m__ }` }`;",
	"var a = `${\"{\"}__m__`;",
	"var a = `${ f({ b: { c: 1 } }) + __m__ }`;",
	"var a = `${ /{/.source }__m__`;",
	"// \"\nvar a = __m__;",
	`/* ' */ var a = __m__;`,
	`/* __m__ */ var a = 1;`,
	`var r = /"/; var a = __m__;`,
	`f(/'/, __m__);`,
	`function f() { return /"/; } var a = __m__;`,
	`if (b) {} /"/.test(s); var a = __m__;`,
	`var a = b / 2; var c = "/"; var d = __m__;`,
	`var a = (b) / 2; var c = "/"; var d = __m__;`,
	`var a = b[0] / 2; var c = "/"; var d = __m__;`,
	`var a = __m__ / 2; var c = "/"; var d = __m__;`,
	`var a = /b/g / 2; var c = "/"; var d = __m__;`,
	`if (ok) /"/.test(s); var a = __m__;`,
	`for (;;) /"/.test(s); var a = __m__;`,
	`async function f() { for await (const a of b) /"/.test(a); } var c = __m__;`,
	`if ((a) / 2) { var c = "/"; } var d = __m__;`,
	`var a = b.if(c) / 2; var d = "/"; var e = __m__;`,
	`var a = b.return / 2; var c = "/"; var d = __m__;`,
	`var a = b?.return / 2; var c = "/"; var d = __m__;`,
	`var a = i++ / 2; var c = "/"; var d = __m__;`,
	`var a = i+++/"/.source; var b = __m__;`,
	`var r = /a__m__/;`,
	`var r = /[/]__m__/;`,
	`var r = /\/__m__/;`,
	`var r = /__m__/g; var a = __m__;`,
	"var a = { b: __m__ }, c = `${ { d: __m__ }.d }__m__`;",
	"class A { #b = /\"/; c() { return this.#b / 2; } } var d = __m__;",
	"var f = () => /\"/; var g = a => a / 2; var h = '/' + __m__;",
	"label: for (const a of /x/g.exec(s)) { break label; } var b = __m__;",
	"var a = typeof /x/; var b = void /\"/; var c = __m__;",
	"var a = x\n/b/g\nvar c = '/' + __m__;",
	"var a = yield\n/b/g; var c = __m__;",
	"var a = { get b() { return 1 } } / 2; var c = __m__;",
	"__m__0.__m__",
	"var café = 1; var a = café/__m__/2; var b = \"/\";",
	"var A\u05B1 = 1; var a = A\u05B1/\"/g, b = __m__ + \"/ 1;",
	"function f() { return\u2028/\"/; } var a = __m__;",
	"var a = b\u2028__m__;",
	"// a\u2028var a = `__m__`;",
	"var r = /a\u2028var b = \"__m__\";",
}

// FuzzTokenizerMatchesParser asserts that the tokenizer classifies each marker in a
// valid JavaScript program the same way as a JavaScript parser does.
func FuzzTokenizerMatchesParser(f *testing.F) {
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		expected, ignoredOffsets, ok := classifyMarkersWithParser(source)
		if !ok {
			t.Skip("the source is not a valid program that the tokenizer is expected to classify")
		}
		actual, ambiguousOffset, isAmbiguous, err := classifyMarkersWithTokenizer(source, ignoredOffsets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isAmbiguous {
			// The templ parser returns an error for the marker at ambiguousOffset, so the
			// markers from there on are not escaped.
			for offset := range expected {
				if offset >= ambiguousOffset {
					delete(expected, offset)
				}
			}
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("source %q: map of marker offset to context mismatch (-parser +tokenizer):\n%s", source, diff)
		}
	})
}

// classifyMarkersWithTokenizer reads source with the tokenizer, in the same way as the
// templ parser reads a script element, and returns the context of each marker that the
// tokenizer reads outside a comment. A marker at one of ignoredOffsets isn't a value
// position, e.g. the marker in __m__0, so it is read as source. If the tokenizer returns
// ErrAmbiguousContext for a marker, it returns the offset of the marker, and true.
func classifyMarkersWithTokenizer(source string, ignoredOffsets map[int]bool) (offsetToContext map[int]jstokenizer.Context, ambiguousOffset int, isAmbiguous bool, err error) {
	offsetToContext = map[int]jstokenizer.Context{}
	pi := parse.NewInput(source)
	tokenizer := jstokenizer.New()
	for {
		peeked, ok := pi.Peek(len(marker))
		if ok && peeked == marker && !ignoredOffsets[pi.Index()] {
			start := pi.Index()
			pi.Take(len(marker))
			context, err := tokenizer.RecordValue(pi, start)
			if errors.Is(err, jstokenizer.ErrAmbiguousContext) {
				return offsetToContext, start, true, nil
			}
			if err != nil {
				return nil, 0, false, err
			}
			offsetToContext[start] = context
			continue
		}
		_, ok, err = tokenizer.ReadComment(pi)
		if err != nil {
			return nil, 0, false, err
		}
		if ok {
			continue
		}
		_, ok, err = tokenizer.Read(pi)
		if err != nil {
			return nil, 0, false, err
		}
		if !ok {
			return offsetToContext, 0, false, nil
		}
	}
}

// classifyMarkersWithParser parses source as a JavaScript program, and returns the
// context of each marker that isn't in a comment. A marker that is part of a longer
// identifier or an escape sequence, e.g. a__m__ or "\__m__", is not a value position,
// and is returned in ignoredOffsets. It returns false if the source is not a valid
// program, or if the source contains code where the tokenizer is documented to differ
// from a parser.
func classifyMarkersWithParser(source string) (offsetToContext map[int]jstokenizer.Context, ignoredOffsets map[int]bool, ok bool) {
	// The parser doesn't support HTML-like comments, which the tokenizer reads.
	if strings.Contains(source, "<!--") || strings.Contains(source, "-->") {
		return nil, nil, false
	}
	// A browser replaces \r\n and \r with \n in an HTML document before the script is
	// parsed, so the tokenizer never reads source that contains \r.
	if strings.Contains(source, "\r") {
		return nil, nil, false
	}
	program, err := parseProgram(source)
	if err != nil {
		return nil, nil, false
	}
	var nodes []any
	collectNodes(reflect.ValueOf(program), map[uintptr]bool{}, &nodes)
	if containsMisplacedPrivateIdentifier(source, nodes) || containsMislocatedTemplateElement(source, nodes) || containsInvalidRegexp(nodes) || legacyOctalPropertyAccess.MatchString(source) {
		return nil, nil, false
	}

	offsetToContext = map[int]jstokenizer.Context{}
	ignoredOffsets = map[int]bool{}
	markerOffsets := findMarkerOffsets(source)
	for offset := range markerOffsets {
		if isOverlapping(markerOffsets, offset) || isEscaped(source, offset) {
			ignoredOffsets[offset] = true
			continue
		}
		context, found := findLiteralContext(source, nodes, offset)
		if found {
			offsetToContext[offset] = context
			continue
		}
		if isParsedWithCommentMarker(source, offset) {
			continue
		}
		if isMarkerIdentifier(source, nodes, offset) {
			offsetToContext[offset] = jstokenizer.ContextCode
			continue
		}
		ignoredOffsets[offset] = true
	}
	return offsetToContext, ignoredOffsets, true
}

// findMarkerOffsets returns the byte offset of each marker in source, including
// markers that overlap, e.g. both markers in __m__m__.
func findMarkerOffsets(source string) (offsets map[int]bool) {
	offsets = map[int]bool{}
	var from int
	for {
		index := strings.Index(source[from:], marker)
		if index < 0 {
			return offsets
		}
		offsets[from+index] = true
		from += index + 1
	}
}

// isOverlapping returns true if the marker at offset overlaps another marker. The
// tokenizer reads only the first of two overlapping markers.
func isOverlapping(markerOffsets map[int]bool, offset int) bool {
	for distance := 1; distance < len(marker); distance++ {
		if markerOffsets[offset-distance] || markerOffsets[offset+distance] {
			return true
		}
	}
	return false
}

// isEscaped returns true if the marker at offset follows an odd number of backslashes,
// so that its first character is part of an escape sequence, e.g. "\__m__".
func isEscaped(source string, offset int) bool {
	var backslashes int
	for i := offset - 1; i >= 0 && source[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

// findLiteralContext returns the context of the string, template literal, or regular
// expression literal that contains offset.
func findLiteralContext(source string, nodes []any, offset int) (context jstokenizer.Context, ok bool) {
	for _, node := range nodes {
		switch node := node.(type) {
		case *ast.StringLiteral:
			// An identifier used as a property name is a string literal without quotes.
			start := int(node.Idx) - 1
			if source[start] != '"' && source[start] != '\'' {
				continue
			}
			if isWithin(offset, start, len(node.Literal)) {
				return jstokenizer.ContextString, true
			}
		case *ast.TemplateElement:
			if isWithin(offset, getTemplateElementStart(node), len(node.Literal)) {
				return jstokenizer.ContextTemplateLiteral, true
			}
		case *ast.RegExpLiteral:
			// The parser's Literal can include a comment that follows the flags, so the
			// pattern is used instead.
			if !isWithin(offset, int(node.Idx), len(node.Pattern)) {
				continue
			}
			if offset == int(node.Idx) {
				return jstokenizer.ContextRegexpStart, true
			}
			return jstokenizer.ContextRegexp, true
		}
	}
	return jstokenizer.ContextCode, false
}

// getTemplateElementStart returns the byte offset of the text of a template literal
// element. The parser's Idx is one byte after it, and one-based.
func getTemplateElementStart(element *ast.TemplateElement) int {
	return int(element.Idx) - 2
}

// isWithin returns true if the marker at offset is within the span of length bytes at
// start.
func isWithin(offset, start, length int) bool {
	return offset >= start && offset+len(marker) <= start+length
}

// isMarkerIdentifier returns true if the marker at offset is a complete identifier,
// or a complete property name.
func isMarkerIdentifier(source string, nodes []any, offset int) bool {
	next, _ := utf8.DecodeRuneInString(source[offset+len(marker):])
	if isIdentifierPart(next) {
		return false
	}
	previous, _ := utf8.DecodeLastRuneInString(source[:offset])
	if isIdentifierPart(previous) {
		return false
	}
	for _, node := range nodes {
		switch node := node.(type) {
		case *ast.Identifier:
			if int(node.Idx)-1 == offset && string(node.Name) == marker {
				return true
			}
		case *ast.StringLiteral:
			if int(node.Idx)-1 == offset && node.Literal == marker {
				return true
			}
		}
	}
	return false
}

// isIdentifierPart returns true if r can be part of an identifier, including the \ of an
// escape sequence, see https://tc39.es/ecma262/#prod-IdentifierPartChar.
func isIdentifierPart(r rune) bool {
	return r == '$' || r == '\\' || r == '\u200c' || r == '\u200d' || unicode.In(r, unicode.L, unicode.Nl, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc)
}

// isParsedWithCommentMarker returns true if source is still a valid program after the
// marker at offset is replaced with commentMarker. The marker isn't in a literal, so it
// must be in a comment.
func isParsedWithCommentMarker(source string, offset int) bool {
	_, err := parseProgram(source[:offset] + commentMarker + source[offset+len(marker):])
	return err == nil
}

// containsMislocatedTemplateElement returns true if the text of a template literal
// element isn't at its position in the source, e.g. because the parser normalizes a
// \r\n line terminator to \n.
func containsMislocatedTemplateElement(source string, nodes []any) bool {
	for _, node := range nodes {
		element, ok := node.(*ast.TemplateElement)
		if !ok {
			continue
		}
		start := getTemplateElementStart(element)
		if start < 0 || start+len(element.Literal) > len(source) || source[start:start+len(element.Literal)] != element.Literal {
			return true
		}
	}
	return false
}

// containsInvalidRegexp returns true if a regular expression literal has a flag that
// isn't valid, or has no closing /. The parser accepts these, e.g. /a/z and /[a, but a
// browser rejects the program.
func containsInvalidRegexp(nodes []any) bool {
	for _, node := range nodes {
		literal, ok := node.(*ast.RegExpLiteral)
		if !ok {
			continue
		}
		if strings.Trim(literal.Flags, "dgimsuvy") != "" || !strings.HasPrefix(literal.Literal, "/"+literal.Pattern+"/"+literal.Flags) {
			return true
		}
	}
	return false
}

// legacyOctalPropertyAccess matches a legacy octal number such as 00, which can't have a
// decimal point, followed by a property access, e.g. `00.x`. ADR 0005 documents that the
// tokenizer reads it as a number with a decimal point.
var legacyOctalPropertyAccess = regexp.MustCompile(`(^|[^\w$.])0[0-7]+\.`)

// collectNodes appends every struct reachable from v to nodes, as a pointer.
func collectNodes(v reflect.Value, visited map[uintptr]bool, nodes *[]any) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() || visited[v.Pointer()] {
			return
		}
		visited[v.Pointer()] = true
		if v.Elem().Kind() == reflect.Struct {
			*nodes = append(*nodes, v.Interface())
			collectStructFields(v.Elem(), visited, nodes)
			return
		}
		collectNodes(v.Elem(), visited, nodes)
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		collectNodes(v.Elem(), visited, nodes)
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			collectNodes(v.Index(i), visited, nodes)
		}
	case reflect.Struct:
		pointer := reflect.New(v.Type())
		pointer.Elem().Set(v)
		*nodes = append(*nodes, pointer.Interface())
		collectStructFields(v, visited, nodes)
	}
}

func collectStructFields(v reflect.Value, visited map[uintptr]bool, nodes *[]any) {
	if v.Type() == reflect.TypeFor[file.File]() {
		return
	}
	for i := range v.NumField() {
		if !v.Type().Field(i).IsExported() {
			continue
		}
		collectNodes(v.Field(i), visited, nodes)
	}
}

// parseProgram parses source as a JavaScript program. The parser panics on some
// invalid programs, e.g. var[...[, so a panic is returned as an error.
func parseProgram(source string) (program *ast.Program, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("parser panicked: %v", r)
		}
	}()
	return parser.ParseFile(nil, "", source, 0)
}

// containsMisplacedPrivateIdentifier returns true if a private identifier, e.g. #a, is
// outside a class, and isn't a property name or the left operand of in. The parser
// accepts a private identifier as an expression statement, e.g. `#a`, which a browser
// rejects.
func containsMisplacedPrivateIdentifier(source string, nodes []any) bool {
	for _, node := range nodes {
		identifier, ok := node.(*ast.PrivateIdentifier)
		if !ok {
			continue
		}
		start := int(identifier.Idx) - 1
		if strings.HasSuffix(strings.TrimSpace(source[:start]), ".") || strings.HasPrefix(strings.TrimSpace(source[findIdentifierEnd(source, start+1):]), "in") {
			continue
		}
		if !isWithinClass(nodes, start) {
			return true
		}
	}
	return false
}

// findIdentifierEnd returns the byte offset of the end of the identifier that starts at
// start. The parser's Idx1 is calculated from the decoded name, which is a different
// length to the source for an identifier that contains a non-ASCII character or an
// escape sequence.
func findIdentifierEnd(source string, start int) (end int) {
	end = start
	for end < len(source) {
		r, size := utf8.DecodeRuneInString(source[end:])
		if !isIdentifierPart(r) {
			return end
		}
		end += size
	}
	return end
}

// isWithinClass returns true if offset is within the source of a class.
func isWithinClass(nodes []any, offset int) bool {
	for _, node := range nodes {
		class, ok := node.(*ast.ClassLiteral)
		if ok && offset >= int(class.Class)-1 && offset < int(class.RightBrace)-1 {
			return true
		}
	}
	return false
}
