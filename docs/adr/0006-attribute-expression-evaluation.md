# ADR 0006: Evaluate attribute expressions once, in place

## Status

Accepted

## Context

Some attribute values write output before the opening tag of their element:

- An event handler, e.g. `onclick={ handler() }`, writes the function of a `templ.ComponentScript` in a `<script>` element, so that the function is defined before the element calls it.
- A `class` expression, e.g. `class={ className() }`, writes the CSS of each `templ.ComponentCSSClass` in a `<style>` element.
- An attribute key expression or spread attribute can contain a `templ.ComponentScript`, so it writes its function in the same way as an event handler (ADR 0004).

The generator wrote this output by copying each expression in front of the opening tag, and then writing the attributes. This had the following defects:

- Each event handler, attribute key expression, and spread attribute expression was evaluated twice: once to write the script, and once to write the attribute. An expression with side effects ran twice, and an expression that returned a different `templ.ComponentScript` on each call wrote the function of the first and called the function of the second, which was not defined.
- Expressions within conditional attributes were copied out of their branch, so `if show { onclick={ handler() } }` evaluated `handler()` when `show` was false, and an `else` branch evaluated both branches. An error from an expression in a branch that was not selected failed the render.
- A `class` expression in a branch that was not selected wrote the CSS of a class that no element used.

## Decision

Evaluate each attribute expression once, where it is written, inside its branch.

For an element with an attribute that writes output before the opening tag, generated code writes the opening tag to a pooled buffer, `templ_7745c5c3_Buffer`, which shadows the writer of the component within a block. The writer of the component is assigned to `templ_7745c5c3_ElementWriter` first. Attribute expressions write scripts and CSS to `templ_7745c5c3_ElementWriter`, and attribute names and values to the buffer. After the attributes, the buffer is written to `templ_7745c5c3_ElementWriter`, so the scripts and CSS come before the opening tag, as before.

- `class` expressions call `templ.RenderCSSItems` with the element writer.
- Event handlers with constant names append their script to `templ_7745c5c3_ElementScripts`, which is rendered with a single call to `templ.RenderScriptItems` before the buffer is written. This keeps the scripts of the event handlers of an element in a single `<script>` element, as before.
- `templruntime.AttributeExpression.ResolveDynamic` and `templruntime.RenderAttributes` take the element writer, and render the function of a `templ.ComponentScript` value as they resolve it.

Elements without these attributes are written directly, as before.

The rendered HTML is unchanged, except that the CSS of a class in a branch that is not selected is no longer written.

## Consequences

- Each attribute expression is evaluated once, and only when its branch is selected.
- The function that is written is always the function that the attribute calls.
- Each element that has a `class` expression, an event handler, an attribute key expression, or spread attributes gets a buffer from a pool, and copies its opening tag. In a benchmark of a list item with a `class` expression and an event handler, rendering took less time than before, because the event handler was no longer evaluated twice.
- The order in which scripts and CSS are written follows the order of the attributes, except that the scripts of event handlers with constant names are written last.

## Alternatives considered

1. **Assign each expression to a variable before the opening tag.** The variable must be declared outside the branch with the type of the expression, which the generator doesn't know, and the conditions of conditional attributes would have to be evaluated before the opening tag too, which evaluates them twice, or moved, which changes the scope of variables declared in them.
2. **Write scripts and CSS after the closing tag of the element.** This avoids a buffer, but changes the rendered HTML of existing templates. When the response is streamed, a `<style>` element after an element applies after the element has been rendered.
