# ADR 0001: Inline templ element expressions after text

## Status

Accepted

## Context

Issue #693 reported that `@Component()` appearing after text on the same line (e.g., `Left: @left()`) was treated as plain text instead of a component call. The text parser consumed the entire line, including the `@component()` call, because it did not recognize `@` as a delimiter when preceded by whitespace within text content.

This made it impossible to write inline component calls after text, such as `Label: @icon() Home`, without restructuring the template to place the component on its own line.

## Decision

`@` preceded by whitespace (space or tab) in a text context starts a templ element expression. The text parser treats `" @"` and `"\t@"` as delimiters, stopping text consumption before the whitespace. The whitespace is consumed as trailing space on the text node, and the `@` begins a new templ element expression.

## Consequences

- `Label: @icon()` parses as `Text("Label:")` followed by `TemplElementExpression("icon()")`, with the space preserved as trailing whitespace on the text node.
- `user@example.com` continues to parse as plain text because there is no whitespace before `@`.
- `<div>@component</div>` continues to work because the element parser's children loop starts fresh at `@`.
- To include a literal `@` after a space (e.g., a social media handle), use a string expression: `{ "@username" }`.

## Alternatives considered

1. **`@` always starts a component call.** This would break email addresses and other text containing `@`, requiring escaping in common cases.
2. **Require `@` on its own line.** This was the previous behavior, which was limiting and unintuitive.
3. **Check if `@identifier` matches a known parameter.** Proposed in the issue discussion, but fragile because it depends on scope analysis and would not work for imported components or method calls.

## Rationale

The whitespace-before-`@` rule is consistent with how `@` already works at the start of a line (preceded by indentation whitespace). It preserves backward compatibility with email addresses and other `@`-containing text, while enabling the intuitive inline syntax that users expect. The escape hatch via `{ "@text" }` handles edge cases where a literal `@` after a space is needed.
