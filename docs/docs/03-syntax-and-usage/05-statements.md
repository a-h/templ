# Statements

## Control Flow

Within a templ element, some Go statements are valid. Control flow statements can be used to render child elements
conditionally or multiple times.

For individual implementation guides see: [If/else](/syntax-and-usage/if-else), [Switch](/syntax-and-usage/switch) and
[For loops](/syntax-and-usage/loops)

## Incomplete Statements

The templ parser does not make any assumptions on whether a templ element is a broken control flow statement or a text
element.

The following would cause the templ parser to return an error to prevent backend template code from being leaked.

```templ title="broken-if.templ"
package main

templ showIfTrue(b bool) {
	if b 
	  <p>Hello</p>
	}
}
```

:::note
Note the missing `{` on line 4.
:::

The following would also error as the parser will not distinguish the sentence from a broken `if`.

```templ title="paragraph.templ"
package main

templ text(b bool) {
	<p>if a tree fell in the woods</p>
}
```

:::note
This also applies to `for` and `switch` statements.
:::

To tell the parser the intention was for it to be a text element and not a control flow, either capitalise the first
letter, or use a Go string literal.

```templ title="paragraph.templ"
package main

templ text(b bool) {
	<p>If a tree fell in the woods</p>
	<p>{ "if a tree fell in the woods" }</p>
}
```
