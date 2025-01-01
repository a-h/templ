# Generating static HTML files with templ

templ components implement the `templ.Component` interface.

The interface has a `Render` method which outputs HTML to an `io.Writer` that is passed in.

```go
type Component interface {
	// Render the template.
	Render(ctx context.Context, w io.Writer) error
}
```

In Go, the `io.Writer` interface is implemented by many built-in types in the standard library, including `os.File` (files), `os.Stdout`, and `http.ResponseWriter` (HTTP responses).

This makes it easy to use templ components in a variety of contexts to generate HTML.

To render static HTML files using templ component, first create a new Go project.

## Setup project

Create a new directory.

```bash
mkdir static-generator
```

Initialize a new Go project within it.

```bash
cd static-generator
go mod init github.com/a-h/templ-examples/static-generator
```

## Create a templ file

To use it, create a `hello.templ` file containing a component.

Components are functions that contain templ elements, markup, `if`, `switch` and `for` Go expressions.

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

Create a `main.go` file. The program creates a `hello.html` file and uses the component to write HTML to the file.

```go title="main.go"
package main

import (
	"context"
	"log"
	"os"
)

func main() {
	f, err := os.Create("hello.html")
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}

	err = hello("John").Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}
}
```

## Run the program

Running the code will create a file called `hello.html` containing the component's HTML.

```bash
go run *.go
```

```html title="hello.html"
<div>Hello, John</div>
```
