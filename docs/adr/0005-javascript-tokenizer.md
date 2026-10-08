# ADR 0005: Classify Go expressions in script elements with a JavaScript tokenizer

## Status

Accepted

## Context

templ escapes each `{{ }}` Go expression in a `<script>` element for its position in the JavaScript source. A value in code is JSON encoded, a value in a single or double quoted string is escaped as string content, and a value in a template literal has `$`, `{`, and `}` escaped as well, so that it can't start a `${ }` substitution (GHSA-jq3r-rjqp-mwgj).

The parser classified each position with a character scanner that was part of the `<script>` element parser. The scanner tracked quotes, comments, and template literal substitutions with separate rules in a single loop, and had the following defects:

- Braces inside strings and comments in a `${ }` substitution were counted as code, so in `` `${"{"}{{ .Value }}` `` the substitution was treated as still open, and the value was JSON encoded rather than escaped for the template literal. A value of `${alert(1)}` then executed.
- Quotes and comments inside a substitution were not tracked, because the scanner remained in template literal state within it.
- A `/` in code was always tried as the start of a regular expression, so in `var a = b / 2; var c = "/"; var d = {{ v }};` the text `/ 2; var c = "/` was read as a regular expression. The scanner then treated `{{ v }}` as being inside a string, and escaped it as string content in a code position, where it executed.
- `//` and `/*` were read as comments after a Go expression inside a string.

## Decision

Move the scanning into a JavaScript tokenizer in `internal/jstokenizer`. The tokenizer reads one token at a time, i.e. a comment, a regular expression, an identifier or keyword, an escape sequence, or a single character, and tracks:

- Whether it is in code, a single or double quoted string, or a template literal.
- A stack of open `${ }` substitutions, each with a count of the unmatched `{` tokens in code since it opened. The contents of a substitution are tokenized as code, so strings, comments, regular expressions, and nested template literals within it are handled in the same way as anywhere else.
- Whether a `/` in code starts a regular expression or is a division operator, based on the previous token. A regular expression can only start an expression, i.e. at the start of the source, after an operator, `(`, `[`, `{`, `}`, `,`, `;`, or a keyword such as `return` or `typeof`. After an identifier, a number, `)`, `]`, a string, a template literal, a regular expression, or an inserted value, `/` is division.

`Tokenizer.GetContext` returns a `jstokenizer.Context` of `ContextCode`, `ContextString`, or `ContextTemplateLiteral`, and the parser maps it to a `parser.ScriptContentsContext`. The parser reads `{{ }}` expressions itself, and calls `Tokenizer.RecordValue` after each, so that a `/` after a Go expression is division. The tokenizer doesn't depend on the parser's types, in the same way that `internal/htmlattr` classifies attributes without depending on the generator or the runtime (ADR 0004).

The package is internal, so that its API is not a compatibility commitment of the public `parser/v2` module.

## Consequences

- Braces in strings, comments, and regular expressions inside a substitution don't affect where the substitution ends.
- A value after a division operator is classified by the code that follows the division, rather than by text that was misread as a regular expression.
- The tokenizer is tested independently of the parser, by inserting a marker value into JavaScript source and asserting the context at the marker.
- `}` followed by `/` is always treated as the start of a regular expression. This is correct after a block, e.g. `if (a) {} /x/.test(s)`, and incorrect after an object literal, e.g. `x = {} / 2`. Dividing an object literal has no practical use.
- `++` and `--` are treated as operators that precede an expression, so in `i++ / 2`, the `/` is misread as the start of a regular expression. Postfix increment followed by division on the same line is rare in templates.
- A `<script>` element without a `</script>` end tag is always a parse error. The previous scanner accepted one when the input ended directly after a character.

## Alternatives considered

1. **Use a JavaScript parser, such as esbuild or goja.** The source isn't valid JavaScript while it contains `{{ }}` expressions, and the escaping only depends on the lexical context of each expression, not on the syntax tree, so a parser would require replacing each expression with a placeholder whose validity depends on its position, and mapping the placeholders back.
2. **Use an existing tokenizer, such as `tdewolff/parse/v2/js`.** This has the same placeholder problem as a parser, adds a dependency, and the tokenizer still has to report the context of each placeholder, which is the part that templ needs.
3. **Fix the brace counting in the existing scanner.** This fixed the reported defect, but left the regular expression and comment defects, and kept the lexical state spread across a single loop in the `<script>` element parser.
