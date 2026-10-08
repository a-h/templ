package runtime

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

// scriptContext describes the JavaScript syntax position that a Go value is
// being written into, which determines how the value must be escaped.
type scriptContext int

const (
	// scriptContextExpression is outside of any string literal, so the value
	// is JSON encoded and used directly as a JavaScript expression.
	scriptContextExpression scriptContext = iota
	// scriptContextString is inside a single or double quoted string literal.
	scriptContextString
	// scriptContextTemplateLiteral is inside a backtick quoted template literal,
	// outside of any ${ ... } substitution.
	scriptContextTemplateLiteral
)

func ScriptContentInsideStringLiteral[T any](v T, errs ...error) (string, error) {
	return scriptContent(v, scriptContextString, errs...)
}

func ScriptContentInsideTemplateLiteral[T any](v T, errs ...error) (string, error) {
	return scriptContent(v, scriptContextTemplateLiteral, errs...)
}

func ScriptContentOutsideStringLiteral[T any](v T, errs ...error) (string, error) {
	return scriptContent(v, scriptContextExpression, errs...)
}

func scriptContent[T any](v T, context scriptContext, errs ...error) (string, error) {
	if errors.Join(errs...) != nil {
		return "", errors.Join(errs...)
	}
	table := jsStrReplacementTable
	if context == scriptContextTemplateLiteral {
		table = jsBqStrReplacementTable
	}
	if vs, ok := any(v).(string); ok && context != scriptContextExpression {
		return replace(vs, table), nil
	}
	jd, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if context != scriptContextExpression {
		return replace(string(jd), table), nil
	}
	return string(jd), nil
}

// See https://cs.opensource.google/go/go/+/refs/tags/go1.23.6:src/html/template/js.go

// replace replaces each rune r of s with replacementTable[r], provided that
// r < len(replacementTable). If replacementTable[r] is the empty string then
// no replacement is made.
// It also replaces runes U+2028 and U+2029 with the raw strings `\u2028` and
// `\u2029`.
func replace(s string, replacementTable []string) string {
	var b strings.Builder
	r, w, written := rune(0), 0, 0
	for i := 0; i < len(s); i += w {
		// See comment in htmlEscaper.
		r, w = utf8.DecodeRuneInString(s[i:])
		var repl string
		switch {
		case int(r) < len(lowUnicodeReplacementTable):
			repl = lowUnicodeReplacementTable[r]
		case int(r) < len(replacementTable) && replacementTable[r] != "":
			repl = replacementTable[r]
		case r == '\u2028':
			repl = `\u2028`
		case r == '\u2029':
			repl = `\u2029`
		default:
			continue
		}
		if written == 0 {
			b.Grow(len(s))
		}
		b.WriteString(s[written:i])
		b.WriteString(repl)
		written = i + w
	}
	if written == 0 {
		return s
	}
	b.WriteString(s[written:])
	return b.String()
}

var lowUnicodeReplacementTable = []string{
	0: `\u0000`, 1: `\u0001`, 2: `\u0002`, 3: `\u0003`, 4: `\u0004`, 5: `\u0005`, 6: `\u0006`,
	'\a': `\u0007`,
	'\b': `\u0008`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\u000b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	0xe:  `\u000e`, 0xf: `\u000f`, 0x10: `\u0010`, 0x11: `\u0011`, 0x12: `\u0012`, 0x13: `\u0013`,
	0x14: `\u0014`, 0x15: `\u0015`, 0x16: `\u0016`, 0x17: `\u0017`, 0x18: `\u0018`, 0x19: `\u0019`,
	0x1a: `\u001a`, 0x1b: `\u001b`, 0x1c: `\u001c`, 0x1d: `\u001d`, 0x1e: `\u001e`, 0x1f: `\u001f`,
}

var jsStrReplacementTable = []string{
	0:    `\u0000`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\u000b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\u0022`,
	'`':  `\u0060`,
	'&':  `\u0026`,
	'\'': `\u0027`,
	'+':  `\u002b`,
	'/':  `\/`,
	'<':  `\u003c`,
	'>':  `\u003e`,
	'\\': `\\`,
}

// jsBqStrReplacementTable escapes the same runes as jsStrReplacementTable, plus
// '$', '{' and '}'. Those three runes are additionally dangerous inside a
// backtick quoted JavaScript template literal, since an unescaped "${ ... }"
// sequence is evaluated as a JavaScript expression rather than treated as
// string data. See GHSA-jq3r-rjqp-mwgj and CVE-2023-24538, the equivalent
// issue in Go's html/template package, which this table mirrors
// (jsBqStrReplacementTable in html/template/js.go).
var jsBqStrReplacementTable = []string{
	0:    `\u0000`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\u000b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\u0022`,
	'`':  `\u0060`,
	'&':  `\u0026`,
	'\'': `\u0027`,
	'+':  `\u002b`,
	'/':  `\/`,
	'<':  `\u003c`,
	'>':  `\u003e`,
	'\\': `\\`,
	'$':  `\u0024`,
	'{':  `\u007b`,
	'}':  `\u007d`,
}
