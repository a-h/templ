# Components

templ Components are markup and code that is compiled into functions that return a `templ.Component` interface by running the `templ generate` command.

Components can contain templ elements that render HTML, text, expressions that output text or include other templates, and branching statements such as `if` and `switch`, and `for` loops.

```templ title="header.templ"
package main

templ headerTemplate(name string) {
  <header data-testid="headerTemplate">
    <h1>{ name }</h1>
  </header>
}
```

The generated code is a Go function that returns a `templ.Component`.

```go title="header_templ.go"
func headerTemplate(name string) templ.Component {
  // Generated contents
}
```

`templ.Component` is an interface that has a `Render` method on it that is used to render the component to an `io.Writer`.

```go
type Component interface {
	Render(ctx context.Context, w io.Writer) error
}
```

:::tip
Since templ produces Go code, you can share templates the same way that you share Go code - by sharing your Go module.

templ follows the same rules as Go. If a `templ` block starts with an uppercase letter, then it is public, otherwise, it is private.

A `templ.Component` may write partial output to the `io.Writer` if it returns an error. If you want to ensure you only get complete output or nothing, write to a buffer first and then write the buffer to an `io.Writer`.
:::

## Code-only components

Since templ Components ultimately implement the `templ.Component` interface, any code that implements the interface can be used in place of a templ component generated from a `*.templ` file.

```go
package main

import (
	"context"
	"io"
	"os"

	"github.com/a-h/templ"
)

func button(text string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<button>"+text+"</button>")
		return err
	})
}

func main() {
	button("Click me").Render(context.Background(), os.Stdout)
}
```

```html title="Output"
<button>
 Click me
</button>
```

:::warning
This code is unsafe! In code-only components, you're responsible for escaping the HTML content yourself, e.g. with the `templ.EscapeString` function.
:::

## Method components

templ components can be returned from methods (functions attached to types).

Go code:

```templ
package main

import "os"

type Data struct {
	message string
}

templ (d Data) Method() {
	<div>{ d.message }</div>
}

func main() {
	d := Data{
		message: "You can implement methods on a type.",
	}
	d.Method().Render(context.Background(), os.Stdout)
}
```

It is also possible to initialize a struct and call its component method inline.

```templ
package main

import "os"

type Data struct {
	message string
}

templ (d Data) Method() {
	<div>{ d.message }</div>
}

templ Message() {
    <div>
        @Data{
            message: "You can implement methods on a type.",
        }.Method()
    </div>
}

func main() {
	Message().Render(context.Background(), os.Stdout)
}
```

