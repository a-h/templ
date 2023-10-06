# For loops

Use the standard Go `for` loop for iteration.

```templ title="component.templ"
package main

templ nameList(items []Item) {
  <ul>
  for _, item := range items {
    <li>{ item.Name }</li>
  }
  </ul>
}
```

```html title="Output"
<ul>
  <li>A</li>
  <li>B</li>
  <li>C</li>
</ul>
```
