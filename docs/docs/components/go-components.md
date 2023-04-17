---
sidebar_position: 2
---

# Go components

Define components in Go code by implementing the `templ.Component` interface.

```go
type Component interface {
	// Render the template.
	Render(ctx context.Context, w io.Writer) error
}
```

To implement a component as code, you need to implement the interface.

In code components, you're responsible for escaping the HTML content yourself.

```go
package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"os"

	"github.com/a-h/templ"
)

func main() {
	ctx := context.Background()
	codeList([]string{"A", "B", "C"}).Render(ctx, os.Stdout)
}

func codeList(items []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		if _, err = io.WriteString(w, "<ol>"); err != nil {
			return
		}
		for _, item := range items {
			if _, err = io.WriteString(w, fmt.Sprintf("<li>%s</li>", html.EscapeString(item))); err != nil {
				return
			}
		}
		if _, err = io.WriteString(w, "</ol>"); err != nil {
			return
		}
		return nil
	})
}
```
