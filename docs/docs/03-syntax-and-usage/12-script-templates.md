# Using JavaScript with templ

## Script tags

Use standard `<script>` tags, and standard HTML attributes to run JavaScript on the client.

```templ
templ body() {
  <script type="text/javascript">
    function handleClick(event) {
      alert(event + ' clicked');
    }
  </script>
  <button onclick="handleClick(this)">Click me</button>
}
```

To pass data from the server to client-side scripts, see [Passing server-side data to scripts](#passing-server-side-data-to-scripts).

## Adding client side behaviours to components

To ensure that a `<script>` tag within a templ component is only rendered once per HTTP response, use a [templ.OnceHandle](18-render-once.md).

Using a `templ.OnceHandle` allows a component to define global client-side scripts that it needs to run without including the scripts multiple times in the response.

The example below also demonstrates applying behaviour that's defined in a multiline script to its sibling element.

```templ title="component.templ"
package main

import "net/http"

var helloHandle = templ.NewOnceHandle()

templ hello(label, name string) {
  // This script is only rendered once per HTTP request.
  @helloHandle.Once() {
    <script type="text/javascript">
      function hello(name) {
        alert('Hello, ' + name + '!');
      }
    </script>
  }
  <div>
    <input type="button" value={ label } data-name={ name }/>
    <script type="text/javascript">
      // To prevent the variables from leaking into the global scope,
      // this script is wrapped in an IIFE (Immediately Invoked Function Expression).
      (() => {
        let scriptElement = document.currentScript;
        let parent = scriptElement.closest('div');
        let nearestButtonWithName = parent.querySelector('input[data-name]');
        nearestButtonWithName.addEventListener('click', function() {
          let name = nearestButtonWithName.getAttribute('data-name');
          hello(name);
        })
      })()
    </script>
  </div>
}

templ page() {
  @hello("Hello User", "user")
  @hello("Hello World", "world")
}

func main() {
  http.Handle("/", templ.Handler(page()))
  http.ListenAndServe("127.0.0.1:8080", nil)
}
```

:::tip
You might find libraries like [surreal](https://github.com/gnat/surreal) useful for reducing boilerplate.

```templ
var helloHandle = templ.NewOnceHandle()
var surrealHandle = templ.NewOnceHandle()

templ hello(label, name string) {
  @helloHandle.Once() {
    <script type="text/javascript">
      function hello(name) {
        alert('Hello, ' + name + '!');
      }
    </script>
  }
  @surrealHandle.Once() {
    <script src="https://cdn.jsdelivr.net/gh/gnat/surreal@3b4572dd0938ce975225ee598a1e7381cb64ffd8/surreal.js"></script>
  }
  <div>
    <input type="button" value={ label } data-name={ name }/>
    <script type="text/javascript">
      // me("-") returns the previous sibling element.
      me("-").addEventListener('click', function() {
        let name = this.getAttribute('data-name');
        hello(name);
      })
    </script>
  </div>
}
```
:::

## Importing scripts

Use standard `<script>` tags to load JavaScript from a URL.

```templ
templ head() {
	<head>
		<script src="https://unpkg.com/lightweight-charts/dist/lightweight-charts.standalone.production.js"></script>
	</head>
}
```

And use the imported JavaScript directly in templ via `<script>` tags.

```templ
templ body() {
	<script>
		const chart = LightweightCharts.createChart(document.body, { width: 400, height: 300 });
		const lineSeries = chart.addLineSeries();
		lineSeries.setData([
				{ time: '2019-04-11', value: 80.01 },
				{ time: '2019-04-12', value: 96.63 },
				{ time: '2019-04-13', value: 76.64 },
				{ time: '2019-04-14', value: 81.89 },
				{ time: '2019-04-15', value: 74.43 },
				{ time: '2019-04-16', value: 80.01 },
				{ time: '2019-04-17', value: 96.63 },
				{ time: '2019-04-18', value: 76.64 },
				{ time: '2019-04-19', value: 81.89 },
				{ time: '2019-04-20', value: 74.43 },
		]);
	</script>
}
```

:::tip
You can use a CDN to serve 3rd party scripts, or serve your own and 3rd party scripts from your server using a `http.FileServer`.

```go
mux := http.NewServeMux()
mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
http.ListenAndServe("localhost:8080", mux)
```
:::

## Passing server-side data to scripts

Pass data from the server to the client by embedding it in the HTML as a JSON object in an attribute or script tag.

### Pass server-side data to the client in a HTML attribute

```templ title="input.templ"
templ body(data any) {
  <button id="alerter" alert-data={ templ.JSONString(data) }>Show alert</button>
}
```

```html title="output.html"
<button id="alerter" alert-data="{&quot;msg&quot;:&quot;Hello, from the attribute data&quot;}">Show alert</button>
```

The data in the attribute can then be accessed from client-side JavaScript.

```javascript
const button = document.getElementById('alerter');
const data = JSON.parse(button.getAttribute('alert-data'));
```

### Pass server-side data to the client in a script element

```templ title="input.templ"
templ body(data any) {
  @templ.JSONScript("id", data)
}
```

```html title="output.html"
<script id="id" type="application/json">{"msg":"Hello, from the script data"}</script>
```

The data in the script tag can then be accessed from client-side JavaScript.

```javascript
const data = JSON.parse(document.getElementById('id').textContent);
```

## Working with NPM projects

https://github.com/a-h/templ/tree/main/examples/typescript contains a TypeScript example that uses `esbuild` to transpile TypeScript into plain JavaScript, along with any required `npm` modules.

After transpilation and bundling, the output JavaScript code can be used in a web page by including a `<script>` tag.

### Creating a TypeScript project

Create a new TypeScript project with `npm`, and install TypeScript and `esbuild` as development dependencies.

```bash
mkdir ts
cd ts
npm init
npm install --save-dev typescript esbuild
```

Create a `src` directory to hold the TypeScript code.

```bash
mkdir src
```

And add a TypeScript file to the `src` directory.

```typescript title="ts/src/index.ts"
function hello() {
  console.log('Hello, from TypeScript');
}
```

### Bundling TypeScript code

Add a script to build the TypeScript code in `index.ts` and copy it to an output directory (in this case `./assets/js/index.js`).

```json title="ts/package.json"
{
  "name": "ts",
  "version": "1.0.0",
  "scripts": {
    "build": "esbuild --bundle --minify --outfile=../assets/js/index.js ./src/index.ts"
  },
  "devDependencies": {
    "esbuild": "0.21.3",
    "typescript": "^5.4.5"
  }
}
```

After running `npm build` in the `ts` directory, the TypeScript code is transpiled into JavaScript and copied to the output directory.

### Using the output JavaScript

The output file `../assets/js/index.js` can then be used in a templ project.

```templ title="components/head.templ"
templ head() {
	<head>
		<script src="/assets/js/index.js"></script>
	</head>
}
```

You will need to configure your Go web server to serve the static content.

```go title="main.go"
func main() {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Serve components.
	data := map[string]any{"msg": "Hello, World!"}
	h := templ.Handler(components.Page(data))
	mux.Handle("/", h)

	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe("localhost:8080", mux)
}
```

## Script templates

:::warning
Script templates are a legacy feature and are not recommended for new projects. Use standard `<script>` tags to import a standalone JavaScript file, optionally created by a bundler like `esbuild`.
:::

If you need to pass Go data to scripts, you can use a script template.

Here, the `page` HTML template includes a `script` element that loads a charting library, which is then used by the `body` element to render some data.

```templ
package main

script graph(data []TimeValue) {
	const chart = LightweightCharts.createChart(document.body, { width: 400, height: 300 });
	const lineSeries = chart.addLineSeries();
	lineSeries.setData(data);
}

templ page(data []TimeValue) {
	<html>
		<head>
			<script src="https://unpkg.com/lightweight-charts/dist/lightweight-charts.standalone.production.js"></script>
		</head>
		<body onload={ graph(data) }></body>
	</html>
}
```

The data is loaded by the backend into the template. This example uses a constant, but it could easily have collected the `[]TimeValue` from a database.

```go title="main.go"
package main

import (
	"fmt"
	"log"
	"net/http"
)

type TimeValue struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

func main() {
	mux := http.NewServeMux()

	// Handle template.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := []TimeValue{
			{Time: "2019-04-11", Value: 80.01},
			{Time: "2019-04-12", Value: 96.63},
			{Time: "2019-04-13", Value: 76.64},
			{Time: "2019-04-14", Value: 81.89},
			{Time: "2019-04-15", Value: 74.43},
			{Time: "2019-04-16", Value: 80.01},
			{Time: "2019-04-17", Value: 96.63},
			{Time: "2019-04-18", Value: 76.64},
			{Time: "2019-04-19", Value: 81.89},
			{Time: "2019-04-20", Value: 74.43},
		}
		page(data).Render(r.Context(), w)
	})

	// Start the server.
	fmt.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}
```

`script` elements are templ Components, so you can also directly render the Javascript function, passing in Go data, using the `@` expression:

```templ
package main

import "fmt"

script printToConsole(content string) {
	console.log(content)
}

templ page(content string) {
	<html>
		<body>
		  @printToConsole(content)
		  @printToConsole(fmt.Sprintf("Again: %s", content))
		</body>
	</html>
}
```

The data passed into the Javascript funtion will be JSON encoded, which then can be used inside the function.

```go title="main.go"
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Handle template.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Format the current time and pass it into our template
		page(time.Now().String()).Render(r.Context(), w)
	})

	// Start the server.
	fmt.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}
```

After building and running the executable, running `curl http://localhost:8080/` would render:

```html title="Output"
<html>
	<body>
		<script type="text/javascript">function __templ_printToConsole_5a85(content){console.log(content)}</script>
		<script type="text/javascript">__templ_printToConsole_5a85("2023-11-11 01:01:40.983381358 +0000 UTC")</script>
		<script type="text/javascript">__templ_printToConsole_5a85("Again: 2023-11-11 01:01:40.983381358 +0000 UTC")</script>
	</body>
</html>
```

The `JSExpression` type is used to pass arbitrary JavaScript expressions to a templ script template.

A common use case is to pass the `event` or `this` objects to an event handler.

```templ
package main

script showButtonWasClicked(event templ.JSExpression) {
	const originalButtonText = event.target.innerText
	event.target.innerText = "I was Clicked!"
	setTimeout(() => event.target.innerText = originalButtonText, 2000)
}

templ page() {
	<html>
		<body>
			<button type="button" onclick={ showButtonWasClicked(templ.JSExpression("event")) }>Click Me</button>
		</body>
	</html>
}
```
