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

Move the scanning into a JavaScript tokenizer in `internal/jstokenizer`. The tokenizer reads one token at a time, i.e. a comment, an identifier or keyword, an escape sequence, or a single character, and tracks:

- Whether it is in code, a single or double quoted string, a template literal, or the pattern of a regular expression literal. Within a regular expression, it tracks whether it is in a `[ ]` character class, where `/` does not end the regular expression.
- A stack of open `${ }` substitutions, each with a count of the unmatched `{` tokens in code since it opened. The contents of a substitution are tokenized as code, so strings, comments, regular expressions, and nested template literals within it are handled in the same way as anywhere else.
- Whether a `/` in code starts a regular expression or is a division operator, based on the previous token. A regular expression can only start an expression, i.e. at the start of the source, after an operator, `(`, `[`, `{`, `}`, `,`, `;`, or a keyword such as `return` or `typeof`. After an identifier, a number, `)`, `]`, a string, a template literal, a regular expression, an inserted value, or a postfix `++` or `--`, `/` is division.
- A stack of open parentheses, each recording whether it opens the condition of an `if`, `while`, `for`, `for await`, or `with` statement. A statement follows the condition, so a `/` after its `)` starts a regular expression, e.g. `if (a) /x/.test(s)`, while a `/` after other parentheses is division, e.g. `(a) / 2`.
- Whether the previous token is `.`, so that a property name such as `return` in `b.return / 2` is not treated as a keyword.
- Whether only whitespace and comments have been read since the start of the line. A classic script treats `<!--` anywhere, and `-->` at the start of a line, as the start of a single line comment, see [HTML-like comments](https://tc39.es/ecma262/#sec-html-like-comments). Whether nothing has been read, where `#!` starts a [hashbang comment](https://tc39.es/ecma262/#sec-hashbang).

A single line comment ends at any JavaScript line terminator: `\n`, `\r`, U+2028, or U+2029. The tokenizer reads UTF-8 encoded runes, so that U+2028 and U+2029 are read as line terminators, rather than as three bytes, the first of which is a Latin-1 letter.

The state is split across four types:

- `lexer` tracks the lexical state, i.e. the quote, the regular expression character class, and the open substitutions, and reads tokens.
- `Tokenizer` reads the source with a `lexer`, and with the alternative lexers that ADR 0009 adds at an ambiguous `/`.
- `slashClassifier` decides whether a `/` in code starts a regular expression, and owns the parenthesis stack, the previous token and keyword, and the keyword lists. Its decision is a heuristic, because it depends on the grammar around the `/`, which a tokenizer doesn't parse. Where the tokens don't determine it, it reports the `/` as ambiguous, see ADR 0009.
- `commentReader` reads comments, and tracks whether the input is at the start of the source or of a line, where `#!` and `-->` start a comment.

The HTML parser of a browser also treats `<!--` in a `<script>` element as the start of escaped script data, even within a JavaScript string or comment. After `<!--` and then a `<script` start tag, a `</script>` end tag doesn't end the element, see [Restrictions for contents of script elements](https://html.spec.whatwg.org/multipage/scripting.html#restrictions-for-contents-of-script-elements). templ would end the element where the browser doesn't, and escape the Go expressions that follow as HTML rather than JavaScript. The parser tracks the state of the HTML tokenizer of the browser over the script content with a transcription of the script data states of the spec, and returns an error for a `<script` start tag after an unclosed `<!--`. The browser also ends the element at `</script` followed by whitespace, `/`, or `>`, in any case, e.g. `</SCRIPT>` or `</script >` within a string, while templ only ends it at `</script>`, so the parser also returns an error for these.

The escapers encode `<` and `>`, so a Go expression can't start `<!--` or `<script`, but it can contain `!`, `-`, letters, and spaces, so it can complete either, e.g. `"<{{ v }}"` with a value of `!--`. The parser returns an error for a Go expression that directly follows `<`, `<!`, or `<!-`, part of a `</script` end tag, or part of a `<script` start tag after an unclosed `<!--`. A Go expression is assumed not to end escaped script data, so the error is returned whenever the value might leave the browser in escaped script data.

`Tokenizer.GetContext` returns a `jstokenizer.Context` of `ContextCode`, `ContextString`, `ContextTemplateLiteral`, or `ContextRegexp`, and the parser maps it to a `parser.ScriptContentsContext`. A value in a regular expression is escaped with `templruntime.ScriptContentInsideRegexp`, which escapes `/`, `\`, and line terminators, so that the value can't end the regular expression, and the characters that have a special meaning in a pattern, so that the value matches itself. JSON encoding doesn't escape `/`, so a value in a regular expression that was JSON encoded could end it and execute. The parser reads `{{ }}` expressions itself, and calls `Tokenizer.RecordValue` after each, so that a `/` after a Go expression is division. The tokenizer doesn't depend on the parser's types, in the same way that `internal/htmlattr` classifies attributes without depending on the generator or the runtime (ADR 0004).

The package is internal, so that its API is not a compatibility commitment of the public `parser/v2` module.

## Consequences

- Braces in strings, comments, and regular expressions inside a substitution don't affect where the substitution ends.
- A value after a division operator is classified by the code that follows the division, rather than by text that was misread as a regular expression.
- The tokenizer is tested independently of the parser, by inserting a marker value into JavaScript source and asserting the context at the marker. The tokenizer is also compared with a JavaScript parser by a differential fuzz test, see ADR 0007.
- A `/` after `}`, `yield`, `await`, or `of`, or after an identifier at the start of a line, could be a division or the start of a regular expression. The tokenizer reads the source both ways, and a Go expression that the two readings give different contexts is a parse error, see ADR 0009.
- A `++` or `--` is a postfix operator if it follows an expression on the same line, and a prefix operator otherwise, so a `/` after a prefix `++` or `--` starts a regular expression, e.g. `++/x/.lastIndex`.
- A legacy octal number followed by a property access, e.g. `00.x`, is read as a number with a decimal point. A legacy octal number can't have a decimal point, and is a syntax error in strict mode and module scripts.
- `<!--` is treated as a comment in module scripts, where it is not one. This is the same as a classic script, and differs from a module script only in code such as `a <!--b`, i.e. `a < !(--b)`, which has no practical use.
- A Go expression directly after `<` without a space, e.g. `a <{{ b }}`, is a parse error.
- A `<script>` element without a `</script>` end tag is always a parse error. The previous scanner accepted one when the input ended directly after a character.

## Alternatives considered

1. **Use a JavaScript parser, such as esbuild or goja.** The source isn't valid JavaScript while it contains `{{ }}` expressions, and the escaping only depends on the lexical context of each expression, not on the syntax tree, so a parser would require replacing each expression with a placeholder whose validity depends on its position, and mapping the placeholders back.
2. **Use an existing tokenizer, such as `tdewolff/parse/v2/js`.** This has the same placeholder problem as a parser, adds a dependency, and the tokenizer still has to report the context of each placeholder, which is the part that templ needs.
3. **Fix the brace counting in the existing scanner.** This fixed the reported defect, but left the regular expression and comment defects, and kept the lexical state spread across a single loop in the `<script>` element parser.
