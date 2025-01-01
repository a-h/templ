# Running your first templ application

Let's update the previous application to serve HTML over HTTP instead of writing it to the terminal.

## Create a web server

Update the `main.go` file.

templ components can be served as a standard HTTP handler using the `templ.Handler` function.

```go title="main.go"
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

Running the code will start a web server on port 3000.

```bash
go run *.go
```

If you run another terminal session and run `curl` you can see the exact HTML that is returned matches the `hello` component, with the name "John".

```bash
curl localhost:3000
```

```html name="Output"
<div>Hello, John</div>
```
