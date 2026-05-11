# ADR 0002: Prettier formatting for HTML element attributes

## Status

Accepted

## Context

Issue #1263: `templ fmt` runs prettier on script and style element contents, but not on HTML element attributes. Prettier plugins like `prettier-plugin-tailwindcss` operate on attribute values (e.g. sorting Tailwind CSS classes in `class="..."`), so they have no effect when formatting templ files. Users expected `templ fmt` to apply these plugins to their class attributes.

The existing prettier integration constructs synthetic HTML documents (wrapping script/style content in proper HTML tags), sends them through prettier, and extracts the formatted content. This pattern works well and could be extended to attributes.

## Decision

Send constant HTML attribute values through prettier by constructing synthetic HTML elements, one per attribute, and parsing the formatted values back.

Each `ConstantAttribute` with a `ConstantAttributeKey` produces a synthetic `<div data-templ-id="N" key="value"></div>` element. All such elements for a single `HTMLTemplate` block are batched into one synthetic HTML document and formatted in a single prettier invocation. After formatting, the output is parsed with `golang.org/x/net/html` and the formatted values are written back to the AST by correlating via the `data-templ-id` attribute.

Only constant attribute values are formatted. Expression attributes (`name={ expr }`), boolean expression attributes (`name?={ expr }`), and spread attributes (`{ map... }`) contain Go code and are not sent to prettier.

## Consequences

- Prettier plugins that operate on attribute values (e.g. `prettier-plugin-tailwindcss`) work with `templ fmt` and the LSP formatter.
- One additional prettier invocation per `HTMLTemplate` block (batching all attributes). This is the same order of magnitude as the existing per-element calls for script/style content.
- Attribute values may be modified by prettier (e.g. class reordering, whitespace normalization). This is the desired behavior since the user has opted into prettier formatting.
- Attribute order in the templ source is preserved. Only values change.
- The `SingleQuote` flag on `ConstantAttribute` is re-evaluated after formatting, since the value may now contain or no longer contain double-quote characters.

## Alternatives considered

1. **Send the entire templ element (with all attributes) as one synthetic HTML element.** This is simpler but breaks when the same attribute key appears in multiple conditional branches (e.g. `class` in both an `if` block and an `else` block), because HTML does not allow duplicate attribute names. Using one synthetic element per attribute avoids this.
2. **Group attributes by element (multiple attributes per synthetic element).** More efficient but requires handling the duplicate-key problem. The per-attribute approach is simpler and the performance difference is negligible since the document is small and prettier's cost is dominated by startup time.
3. **Run prettier per attribute individually.** Correct but prohibitively slow due to prettier's startup cost per invocation. Batching into one document avoids this.
4. **Use a Go-native Tailwind class sorter instead of prettier.** This would only handle one specific plugin. Using prettier is general-purpose and supports any prettier plugin that processes attributes.
