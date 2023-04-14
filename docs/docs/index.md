---
sidebar_position: 1
---

# Introduction

## Overview of templ

templ is a HTML templating language similar to JSX, designed for writing HTML frontends in Go. It offers a way to create components and manage the structure of an application's frontend with Go code, while also allowing the integration of HTML markup and Go expressions.

Some of the core features and concepts of templ include:

* Components: Reusable pieces of UI that are represented as functions. These functions can take arguments, which allow the customization and configuration of components.
* Parameters: The arguments passed to component functions can be used to display data, or manage conditional rendering, much like React props.
* Conditional rendering: Displaying components based on certain conditions using Go's control structures.
* Rendering lists: Using loops to render multiple components.
* Code generation: Developers create templates, then use the `templ generate` command to generate performant Go code.
* Server-side rendering: Generating HTML on the server side using templ components.
* Static rendering: Generating static HTML files using templ components.

templ enables developers to use the power and flexibility of Go to create robust and maintainable web applications, taking advantage of Go's features, such as its strong typing, concurrency support, and excellent performance.

## Hello World!

## Features and benefits

`todo`

```html
templ PersonTemplate(p Person) {
	<div>
	    for _, v := range p.Addresses {
		    {! AddressTemplate(v) }
	    }
	</div>
}
```

```go
templ PersonTemplate(p Person) {
	<div>
	    for _, v := range p.Addresses {
		    {! AddressTemplate(v) }
	    }
	</div>
}
```
