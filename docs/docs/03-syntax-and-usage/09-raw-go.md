# Raw Go

:::caution
This page describes functionality that is experimental, and not enabled by default.
To enable this feature run the generation step with the `rawgo` experiment flag: `TEMPLEXPERIMENT=rawgo templ generate`
:::

For some more advanced usecases it may be useful to write go code statements in your template.
Use the `{{ ... }}` syntax to allow for this.

## Variable Declarations

Scoped variables can be created using this syntax, to reduce the need for multiple calls.

```templ title="component.templ"
package main

templ nameList(items []Item) {
    {{ first := items[0] }}
    <p>
        { first.Name }
    </p>
}
```

```html title="Output"
<p>A</p>
```
