# Exporting and sharing

Since templ produces Go code, you can share templates the same way that you share Go code - by sharing your Go module.

templ follows the same rules as Go. If a `templ` block starts with an uppercase letter, then it is public, otherwise, it is private.

```html title="greeting.templ"
package main

// Private.
templ hello(name string) {
  <div>Hello, { name }</div>
}

// Public.
templ Greeting(person Person) {
  <div class="greeting">
    @hello(person.Name)
  </div>
}
```
