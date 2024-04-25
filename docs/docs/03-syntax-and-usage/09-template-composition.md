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

Components can also be passed as parameters and rendered using the `@component` expression.

```templ
package main

templ heading() {
    <h1>Heading</h1>
}

templ layout(contents templ.Component) {
	<div id="heading">
		@heading()
	</div>
	<div id="contents">
		@contents
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
	c := paragraph("Dynamic contents")
	layout(c).Render(context.Background(), os.Stdout)
}
```

```html title="output"
<div id="heading">
	<h1>Heading</h1>
</div>
<div id="contents">
	<p>Dynamic contents</p>
</div>
```

You can pass `templ` components as parameters to other components within templates using standard Go function call syntax.

```templ
package main

templ heading() {
    <h1>Heading</h1>
}

templ layout(contents templ.Component) {
	<div id="heading">
		@heading()
	</div>
	<div id="contents">
		@contents
	</div>
}

templ paragraph(contents string) {
	<p>{ contents }</p>
}

templ root() {
	@layout(paragraph("Dynamic contents"))
}
```

```go title="main.go"
package main

import (
	"context"
	"os"
)

func main() {
	root().Render(context.Background(), os.Stdout)
}
```

```html title="output"
<div id="heading">
	<h1>Heading</h1>
</div>
<div id="contents">
	<p>Dynamic contents</p>
</div>
```

# Referencing Components Across Different Packages

When working on a project organized into multiple packages, it's often necessary to reference components from other packages to maintain clean and efficient code organization. Consider the following project structure:

#### Project structure example

```
my-project/
├── ui/
│   ├── pages/
│   │   ├── home.templ
│   │   ├── faq.templ
│   ├── components/
|       ├── button.templ
│   ├── layout.templ
```

#### Project components

In the `layout.templ` file, we declare it as part of the `ui` package. It contains a `Layout` component, which provides a basic HTML structure for every page of the site:

```templ title="my-project/ui/layout.templ"
package ui

templ Layout() {
    <html>
        <head>
            <title>My Project</title>
        </head>
        <body>
            <main>
                { children... }
            </main>
        </body>
    </html>
}
```

From within the `home.templ` file, you can import and use the `Layout` component from the `ui` package:

```templ title="my-project/ui/pages/home.templ"
package pages

import (
	"my-project/ui"
)

templ Home() {
    // Import the `Layout` component from the `ui` package for use here
    @ui.Layout() {
        <h1>Home Page</h1>
    }
}
```

This method allows for great project organization and component reuse. For example, consider a reusable button component:

```templ title="my-project/ui/components/button.templ"
package components

templ Button(title string) {
    <button class="my-button">{ title }</button>
}
```

You can now reuse this button component across various parts of your project:

```templ title="my-project/ui/pages/home.templ"
package home

import (
    "my-project/ui"
    "my-project/ui/components"
)

templ Home() {
    @ui.Layout() {
        <h1>Home Page</h1>
        @components.Button("Click Me!")
    }
}
```

```templ title="my-project/ui/pages/faq.templ"
package pages

import (
    "my-project/ui"
    "my-project/ui/components"
)

templ Faq() {
    @ui.Layout() {
        <p>FAQ</p>
        @components.Button("Know More")
    }
}
```
