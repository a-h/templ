# Creating a simple templ component

To create a templ component, first create a new Go project.

## Setup project

Create a new directory containing our project.

```bash
mkdir hello-world
```

Initialize a new Go project within it.

```bash
cd hello-world
go mod init github.com/a-h/templ-examples/hello-world
go get github.com/a-h/templ
```

## Create a templ file

To use it, create a `hello.templ` file containing a component.

Components are functions that contain templ elements, markup, and `if`, `switch`, and `for` Go expressions.

```templ title="hello.templ"
package main

templ hello(name string) {
	<div>Hello, { name }</div>
}
```

## Generate Go code from the templ file

Run the `templ generate` command.

```bash
templ generate
```

templ will generate a `hello_templ.go` file containing Go code.

This file will contain a function called `hello` which takes `name` as an argument, and returns a `templ.Component` that renders HTML.

```go
func hello(name string) templ.Component {
  // ...
}
```

## Write a program that renders to stdout

Create a `main.go` file.

```go title="main.go"
package main

import (
	"context"
	"os"
)

func main() {
	component := hello("John")
	component.Render(context.Background(), os.Stdout)
}
```

## Run the program

Running the code will render the component's HTML to stdout.

```bash
go run .
```

```html title="Output"
<div>Hello, John</div>
```

Instead of passing `os.Stdout` to the component's render function, you can pass any type that implements the `io.Writer` interface. This includes files, `bytes.Buffer`, and HTTP responses.

In this way, templ can be used to generate HTML files that can be hosted as static content in an S3 bucket, Google Cloud Storage, or used to generate HTML that is fed into PDF conversion processes, or sent via email.
