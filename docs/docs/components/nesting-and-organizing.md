# Nesting and organizing

Once you have run `templ generate` in your project, you will see `<name>_templ.go` files within your project that contain generated Go code.

Each `templ` block is converted into a Go function that return a `templ.Component`.

It is possible to nest templates, by using the "call" syntax.

```templ title="greeting.templ"
package main

templ Hello(name string) {
  <div>Hello, { name }</div>
}

templ Greeting(person Person) {
  <div class="greeting">
    // highlight-next-line
    {! Hello(person.Name) }
  </div>
}
```
