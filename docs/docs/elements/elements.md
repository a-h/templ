---
sidebar_position: 1
---

# Elements

templ elements are used to render HTML.

## Output is minified

templ automatically minifies HTML reponses. Given a templ element containing whitespace, templ will strip any whitespace that is not required.

```templ
package main

templ component() {
	<p>
		Text content
	</p>
}
```

```html
<p>Text content</p>
```

## Tags must be closed

templ requires that all HTML elements are closed with either a closing tag (`</a>`), or by using a self-closing element (`<hr/>`).

templ is aware of which HTML elements are "void", and will omit the closing `/` from the element.

```templ
templ component() {
	<div>Test</div>
	<img src="images/test.png"/>
	<br/>
}
```

The unminified output is:

```templ
<div>Test</div>
<img src="images/test.png">
<br/>
```

## Attributes and elements can contain expressions

templ elements can contain placeholder expressions for attributes and content.

```templ
templ button(name string, content string) {
	<button value={ name }>{ content }</button>
}
```

Rendering the component to stdout, we can see the results.

```go
func main() {
	component := button("John", "Say Hello")
	component.Render(context.Background(), os.Stdout)
}
```

```html
<button value="John">Say Hello</button>
```

