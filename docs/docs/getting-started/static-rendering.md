# Static rendering

templ can be used to render HTML for use with static websites.

## Setup project

First, create a new directory containing our project.

```sh
mkdir static-rendering
```

Then initialize a new Go project within it.

```sh
cd static-rendering
go mod init github.com/a-h/templ-examples/static-rendering
```

## Create a templ file

To use it, create a `hello.templ` file containing a component.

```html
package main

templ hello(name string) {
	<div>Hello, { name }</div>
}
```

## Generate code

Run the `templ generate` command.

```sh
templ generate
```

templ will generate a `hello_templ.go` file containing Go code.

This file will contain a function called `hello` which takes `name` as an argument, and returns a `templ.Component` that renders HTML.

```go
func hello(name string) templ.Component {
  // ...
}
```

## Create a web server

Create a `main.go` file.

```go
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

```sh
go run *.go
```

```html
<div>Hello, John</div>
```

Instead of passing `os.Stdout` to the component's render function, you can pass any type that implements the `io.Writer` interface. This includes files, HTTP requests, and HTTP responses.

In this way, templ can be used to generate HTML files that can be hosted as static content in an S3 bucket, Google Cloud Storage, or used to generate HTML that is fed into PDF conversion processes, or sent via email.
