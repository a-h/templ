# For loops

templ supports iterating over slices, arrays and channels etc. using the standard Go `for` loop.

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
