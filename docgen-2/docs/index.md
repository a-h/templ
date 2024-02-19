---
sidebar_position: 1
---

# Introduction

## templ - build HTML with Go

Create components that render fragments of HTML and compose them to create screens, pages, documents, or apps.

* Server-side rendering: Deploy as a serverless function, Docker container, or standard Go program.
* Static rendering: Create static HTML files to deploy however you choose.
* Compiled code: Components are compiled into performant Go code.
* Use Go: Call any Go code, and use standard `if`, `switch`, and `for` statements.
* No JavaScript: Does not require any client or server-side JavaScript.
* Great developer experience: Ships with IDE autocompletion.

```templ
package main

templ Hello(name string) {
  <div>Hello, { name }</div>
}

templ Greeting(person Person) {
  <div class="greeting">
    @Hello(person.Name)
  </div>
}
```
