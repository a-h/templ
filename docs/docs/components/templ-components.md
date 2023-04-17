---
sidebar_position: 1
---

# templ components

In a `.templ` file, you can mix HTML markup and code.

You start a component with the `templ` keyword, followed by the function signature.

Inside the function body, you can add text, elements, and use standard Go `if`, `for` and `switch` statements.

```html
package main

templ list(items []string) {
	<ol>
		for _, item := range items {
			<li>{ item }</li>
		}
	</ol>
}
```

Running the `templ generate` command will create a corresponding `.go` file containing code that renders the component.

In a `main.go` file, you can then use the list component.

```
package main

import (
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	list([]string{"a", "b", "c"}).Render(ctx, os.Stdout)
}
```

```
go run *.go | htmlformat
```

```html
<ol>
 <li>
  a
 </li>
 <li>
  b
 </li>
 <li>
  c
 </li>
</ol>
```

