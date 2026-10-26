# ADR 0009: Return a parse error for a Go expression whose JavaScript context is ambiguous

## Status

Accepted

## Context

The JavaScript tokenizer (ADR 0005) decides whether a `/` in code is a division or the start of a regular expression from the tokens before it. A JavaScript parser decides it from the grammar, which the tokenizer doesn't parse. Where the tokens don't determine it, the tokenizer guessed, and a wrong guess changes how the rest of the source is read. For example, in `if (ok) {} /"/.test(s); var a = {{ x }};`, reading the `/` as a division makes `"/.test(s); var a = ` a string, so `{{ x }}` was escaped as string content in code, where it executes.

The differential fuzz test (ADR 0007) found guesses that are wrong in valid JavaScript:

- `yield`, `await`, and `of` are keywords in some contexts, and identifiers in others, e.g. `var yield = 2; a = yield / {{ x }}`.
- Automatic semicolon insertion ends a statement after an identifier at a line terminator in some statements, e.g. `var a\n/"/.test(s)`, where the `/` starts a regular expression.

ADR 0005 also documented `x = {} / 2` as read incorrectly, because `}` followed by `/` was always read as the end of a block.

Each of these could be fixed with a further heuristic, but each heuristic can be wrong in a case that hasn't been found, and a wrong guess can allow script injection.

## Decision

Where the tokens before a `/` don't determine whether it is a division, the tokenizer reads the source both ways, and returns `ErrAmbiguousContext` for a Go expression if the two readings give it different contexts. The parser reports this as a parse error, which suggests how to rewrite the template.

A `/` is ambiguous after:

- `}`, which ends a block, or an object literal, function expression, or class expression.
- `yield`, `await`, or `of`, unless it is a property name, e.g. `a.yield`.
- An identifier that is followed by a line terminator, or a comment that contains one, unless it is a property name or a keyword.

Other cases are determined by the tokens, so they aren't ambiguous. A `++` or `--` is a postfix operator if it follows an expression on the same line, and a prefix operator otherwise, so a `/` after it is a division or the start of a regular expression respectively. This replaces the rule in ADR 0005 that always read a `/` after `++` or `--` as a division.

The state of the tokenizer for one reading of the source is held by a `lexer`. At an ambiguous `/`, the lexer reads the `/` one way, and returns a copy of itself that has read it the other way. `Tokenizer` keeps the lexer that reads at the position of the parser's input, and a list of the alternative lexers, each with the index of the input that it has read up to. The alternatives read in tokens of different lengths, e.g. one reads a comment that the other reads as code, so each is advanced to the position of the main lexer after each token. When an alternative is at the same position as the main lexer, or as another alternative, and has the same state, it is removed, because it reads the rest of the source in the same way.

For a Go expression, each alternative must be at the start of the expression, i.e. not part way through a longer token such as a comment, and must have the same context as the main lexer. Otherwise, the tokenizer returns `ErrAmbiguousContext`.

A single or double quoted string that contains a line feed or a carriage return is a syntax error, so the lexer ends the string there, as it does for a regular expression. An invalid reading then ends at the next line, and usually has the same state as the valid one, so a Go expression on a later line is accepted.

The number of alternatives is limited to 16. If an ambiguous `/` would exceed it, the context of every later Go expression in the element is treated as ambiguous.

## Consequences

- A Go expression is never escaped for a context that the tokenizer guessed.
- Some valid templates are rejected, e.g. `if (ok) {} /"/.test(s); var a = {{ x }};`. The error message suggests a semicolon before a regular expression, parentheses around a division, or moving the Go expression to a later line. A `/` after `}`, `yield`, `await`, or `of`, or at the start of a line after an identifier, is uncommon in `<script>` elements.
- A Go expression is accepted after an ambiguous `/` once the two readings have the same state, e.g. on the next line in `var a = {} / 2;\nvar b = {{ x }};`.
- The differential fuzz test treats `ErrAmbiguousContext` as a correct result for the Go expression and those after it, because the parser returns an error for them. The fuzz test no longer skips the cases above.
- Each alternative adds the cost of reading the source again until it is removed. Alternatives are only created at an ambiguous `/`, and are usually removed at the next line.

## Alternatives considered

1. **Add a heuristic for each case.** For example, track whether a `}` closes a block by recording the token before each `{`, and track whether the enclosing function is a generator for `yield`. Each heuristic adds state to the tokenizer, and each remaining gap is a potential injection that is only found when someone thinks of it, or the fuzz test does.
2. **Treat the source as ambiguous from an ambiguous `/` until the next `;` or line terminator in code.** This is simpler, but a wrong reading can open a string or a template literal that continues past the next `;` or line terminator, so it doesn't return to the correct state there. Reading the source both ways shows where the two readings agree again.
3. **Use a JavaScript parser.** ADR 0005 rejected this, because a template with `{{ }}` expressions isn't valid JavaScript.
4. **Remove support for Go expressions in `<script>` elements.** The feature is documented and in use. The documentation recommends passing data through attributes or `templ.JSONScript` instead.
