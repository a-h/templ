---
sidebar_position: 4
---

# For loops

templ supports iterating over slices, arrays and channels etc. using the for loop:

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
