# Elements

templ elements are used to render HTML within templ components.

```templ title="button.templ"
package main

templ button(text string) {
	<button class="button">{ text }</button>
}
```

```go title="main.go"
package main

import (
	"context"
	"os"
)

func main() {
	button("Click me").Render(context.Background(), os.Stdout)
}
```

```html title="Output"
<button class="button">
 Click me
</button>
```

:::info
templ automatically minifies HTML responses, output is shown formatted for readability.
:::

## Tags must be closed

Unlike HTML, templ requires that all HTML elements are closed with either a closing tag (`</a>`), or by using a self-closing element (`<hr/>`).

templ is aware of which HTML elements are "void", and will not include the closing `/` in the output HTML.

```templ title="button.templ"
package main

templ component() {
	<div>Test</div>
	<img src="images/test.png"/>
	<br/>
}
```

```templ title="Output"
<div>Test</div>
<img src="images/test.png">
<br>
```

## Attributes and elements can contain expressions

templ elements can contain placeholder expressions for attributes and content.

```templ title="button.templ"
package main

templ button(name string, content string) {
	<button value={ name }>{ content }</button>
}
```

Rendering the component to stdout, we can see the results.

```go title="main.go"
func main() {
	component := button("John", "Say Hello")
	component.Render(context.Background(), os.Stdout)
}
```

```html title="Output"
<button value="John">Say Hello</button>
```
