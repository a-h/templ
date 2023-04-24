---
sidebar_position: 4
---

# Loops

## for

templ supports iterating over slices, arrays and channels etc. using the Go `for` loop.

```templ
templ listNames(items []Item) {
  <ul>
  for _, item := range items {
    <li>{ item.Name }</li>
  }
  </ul>
}
```

```html
<ul>
  <li>A</li>
  <li>B</li>
  <li>C</li>
</ul>
```
