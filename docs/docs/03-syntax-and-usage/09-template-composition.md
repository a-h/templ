# Template composition

Templates can be composed using the import expression.

```templ
templ showAll() {
	@left()
	@middle()
	@right()
}

templ left() {
	<div>Left</div>
}

templ middle() {
	<div>Middle</div>
}

templ right() {
	<div>Right</div>
}
```

```html title="Output"
<div>
 Left
</div>
<div>
 Middle
</div>
<div>
 Right
</div>
```

# Children

Children can be passed to a component for it to wrap.

```templ
templ showAll() {
	@wrapChildren() {
		<div>Inserted from the top</div>
	}
}

templ wrapChildren() {
	<div id="wrapper">
		{ children... }
	</div>
}
```

:::note
The use of the `{ children... }` expression in the child component.
:::

```html title="output"
<div id="wrapper">
 <div>
  Inserted from the top
 </div>
</div>
```

# Components as parameters

Components can also be passed as parameters and rendered using the `{! component }` expression.

```templ
package main

templ layout(l, r templ.Component) {
	<div id="left">
		{! l }
	</div>
	<div id="right">
		{! r }
	</div>
}

templ paragraph(contents string) {
	<p>{ contents }</p>
}
```

```go title="main.go"
package main

import (
	"context"
	"os"
)

func main() {
	l := paragraph("Left contents")
	r := paragraph("Right contents")
	layout(l, r).Render(context.Background(), os.Stdout)
}
```

```html title="output"
<div id="left">
	<p>Left contents</p>
</div>
<div id="right">
	<p>Right contents</p>
</div>
```
