# Raw Go

:::caution
This page describes functionality that is experimental, not enabled by default, and may change or be removed in future versions.

To enable this feature run the generation step with the `rawgo` experiment flag: `TEMPL_EXPERIMENT=rawgo templ generate`

You will also need to set the `TEMPL_EXPERIMENT=rawgo` environment variable at your system level or within your editor to enable LSP behavior.
:::

For some more advanced use cases it may be useful to write Go code statements in your template.

Use the `{{ ... }}` syntax for this.

## Variable declarations

Scoped variables can be created using this syntax, to reduce the need for multiple function calls.

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
