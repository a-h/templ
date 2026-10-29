# Rendering raw HTML

To render HTML that has come from a trusted source, bypassing all HTML escaping and security mechanisms that templ includes, use the `templ.Raw` function.

:::info
Only include HTML that comes from a trusted source.
:::

:::warning
Use of this function may introduce security vulnerabilities to your program.
:::

```templ title="component.templ"
templ Example() {
	<!DOCTYPE html>
	<html>
		<body>
			@templ.Raw("<div>Hello, World!</div>")
		</body>
	</html>
}
```

```html title="Output"
<!DOCTYPE html>
<html>
	<body>
		<div>Hello, World!</div>
	</body>
</html>
```

## Escaping untrusted content inside raw HTML

If raw HTML includes values from an untrusted source, such as user input, escape those values with the `html.EscapeString` function from the Go standard library before you combine them with the HTML.

```templ title="component.templ"
package main

import "html"

templ Greeting(name string) {
	@templ.Raw("<b>" + html.EscapeString(name) + "</b>")
}
```

When `name` is `<script>alert('hello!')</script>`, the component renders:

```html title="Output"
<b>&lt;script&gt;alert(&#39;hello!&#39;)&lt;/script&gt;</b>
```

If the output doesn't need raw HTML, use a string expression such as `<b>{ name }</b>` instead, because templ escapes string expressions automatically. See [Expressions](/syntax-and-usage/expressions#escaping).
