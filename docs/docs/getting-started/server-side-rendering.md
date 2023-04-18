# Server-side rendering

templ ships with a HTTP handler that renders templ components.

## Setup project

First, create a new directory containing our project.

```sh
mkdir server-side
```

Then initialize a new Go project within it.

```sh
cd server-side
go mod init github.com/a-h/templ-examples/server-side
```

## Create a templ file

To use it, create a `hello.templ` file containing a component.

```templ
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
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	component := hello("John")

	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
```

## Run the program

Running the code will start a web server on port 3000 which will return HTML.

```sh
go run *.go
```

If you run another terminal session and run `curl` you can see the exact HTML that is returned matches the `hello` component, with the name "John".

```sh
curl localhost:3000
```

```html
<div>Hello, John</div>
```
