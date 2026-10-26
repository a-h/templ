# ADR 0007: Compare the JavaScript tokenizer with a JavaScript parser in a differential fuzz test

## Status

Accepted

## Context

The JavaScript tokenizer (ADR 0005) classifies the position of each `{{ }}` Go expression in a `<script>` element as code, a string, a template literal, or a regular expression. A wrong classification escapes a value for the wrong context, which can allow script injection.

The tokenizer can't parse JavaScript, so it decides whether a `/` starts a regular expression or is a division with heuristics. Each escaping defect found after the tokenizer was introduced was a case that the heuristics or the lexical rules didn't cover, e.g. `if (ok) /"/`, `b.return / 2`, `i++ / 2`, and `for await (...)`. These were found by reviewers thinking of cases, and the unit tests only cover the cases that someone has thought of.

## Decision

Add a differential fuzz test, `FuzzTokenizerMatchesParser`, which compares the tokenizer with the parser of `github.com/dop251/goja`, a JavaScript engine written in Go.

- The fuzz input is JavaScript source containing the marker `__m__`, which stands in for a `{{ }}` expression. The marker is an identifier in code, and text in a string, template literal, regular expression, or comment, so that a source with markers can be valid JavaScript.
- The test parses the source with goja. If the source isn't a valid program, the input is skipped. Otherwise, the test classifies each marker from the syntax tree: within a string literal, a template literal element, or a regular expression pattern, as an identifier in code, or within a comment, if the source is still valid with the marker replaced by `@@@@@`, which is a syntax error in code.
- The test reads the source with the tokenizer in the same way as the templ parser reads a `<script>` element, and records the context at each marker.
- The test fails if the classifications differ.

ADR 0005 rejected a JavaScript parser at runtime, because a template with `{{ }}` expressions isn't valid JavaScript. That objection doesn't apply here, because the fuzz test only uses inputs that are valid JavaScript.

The test is in a separate module, `internal/jstokenizer/fuzzing`, in the same way as `runtime/fuzzing`, so that goja is not a dependency of templ. The `test-fuzz` task fuzzes it, and the `test-fuzz-corpus` task, which CI runs, runs its seeds and committed corpus.

goja's parser accepts some invalid programs, and reports some positions that don't match the source. The test skips or corrects for these:

- `TemplateElement.Idx` is one byte after the text of the element.
- `RegExpLiteral.Literal` can include a comment after the flags, so the pattern is used.
- The end of an identifier that contains a non-ASCII character or an escape sequence is calculated from the length of its decoded name, so the test finds the end of an identifier by reading the source.
- Invalid regular expression flags, unterminated regular expressions that end in a character class, and private identifiers outside a class are accepted, so these programs are skipped.
- goja panics on some invalid programs, e.g. `var[...[`, so a panic is treated as a parse error.

goja doesn't support HTML-like comments, so sources that contain `<!--` or `-->` are skipped. A browser replaces `\r\n` and `\r` with `\n` in an HTML document before a script is parsed, so sources that contain `\r` are skipped.

## Consequences

The fuzz test found the following, which are fixed:

- U+2028 and U+2029 end a single line comment in JavaScript, but not in the tokenizer. In `// a<U+2028>var s = \`<newline>{{ x }}\``, the tokenizer read the value as code, and JSON encoded it into a template literal, where `${alert(1)}` executes.
- The tokenizer read input one byte at a time, so it read the first byte of U+2028 as the letter `â`, e.g. `return<U+2028>/"/` was read as a division.
- `#!` at the start of the source starts a hashbang comment.
- The numeric literal parser read `_.x` as a number, because it accepted `_` as the first character.
- The numeric literal parser read `0.x` in `.0.x` as a number, because `.0` was read as `.` and `0`.
- The tokenizer only read letters, digits, `_`, and `$` as part of an identifier, so it read a combining mark, e.g. U+05B1 in `A\u05B1`, as punctuation, after which a `/` starts a regular expression. In `a = A\u05B1/"/g, b = {{ x }} + "/ 1`, the value is in a string, but the tokenizer read it as code. The test had skipped programs with non-ASCII identifiers because of the parser's identifier positions, which hid this until the test read the positions from the source instead.

The fuzz test found the following, which depend on the grammar around a `/`, which the tokenizer approximates. ADR 0009 makes a Go expression whose context depends on them a parse error:

- `yield`, `await`, and `of` are identifiers outside a generator, an async function, and a `for ... of` statement respectively. In `var yield = 2; a = yield / {{ x }}`, the tokenizer read a regular expression, so a value that follows is escaped as regular expression content in code.
- Automatic semicolon insertion after a declaration without an initializer, or after `break` or `continue` with a label, ends the statement at a line terminator, so `var a\n/"/.test(s)` starts a regular expression on the second line. The tokenizer read a division after the identifier.

Fuzzing takes time and isn't exhaustive, so the absence of a failure doesn't show that the tokenizer is correct.

## Alternatives considered

1. **Use a generator that produces programs from a JavaScript grammar.** This produces valid programs more often than mutating bytes, but the generator knows the context of each marker only from its own grammar, which can differ from a real parser's, e.g. after automatic semicolon insertion. A parser is still needed as an oracle.
2. **Use V8, through `rogchap.com/v8go`, as the oracle, as `runtime/fuzzing` does.** V8 executes scripts, but doesn't expose a syntax tree or token positions, so it can't report the context of a marker.
3. **Use `tdewolff/parse/v2/js` or esbuild as the oracle.** Either would work. goja's syntax tree includes the position of each literal, which the test needs, and it is written in Go.
