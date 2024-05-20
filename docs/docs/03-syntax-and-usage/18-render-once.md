# Render once

If you need to render something to the page once per page, you can use the `templ.Once` function.

## Example

The `hello` JavaScript function is only rendered once, even though the `hello` component is rendered twice.

```templ title="component.templ"
package once

templ hello(label, name string) {
	@templ.Once("github.com/a-h/templ#hello_script") {
		<script type="text/javascript">
			function hello(name) {
				alert('Hello, ' + name + '!');
			}
		</script>
	}
	<input type="button" value={ label } data-name={ name } onclick="hello(this.getAttribute('data-name'))"/>
}

templ page() {
	@hello("Hello User", "user")
	@hello("Hello World", "world")
}
```

```html title="Output"
<script type="text/javascript">
  function hello(name) {
    alert('Hello, ' + name + '!');
  }
</script>
<input type="button" value="Hello User" data-name="user" onclick="hello(this.getAttribute('data-name'))">
<input type="button" value="Hello World" data-name="world" onclick="hello(this.getAttribute('data-name'))">
```

:::tip
Note the use of the `data-name` attribute to pass the `name` value from server-side Go code to the client-side JavaScript code.

The value of `name` is collected by the `onclick` handler, and passed to the `hello` function.

To pass complex data structures, consider using a `data-` attribute to pass a JSON string using the `templ.JSONString` function, or use the `templ.JSONScript` function to create a templ component that creates a `<script>` element containing JSON data.
:::

## Common use cases

- Rendering a `<style>` tag that contains CSS classes required by a component.
- Rendering a `<script>` tag that contains JavaScript required by a component.
- Rendering a `<link>` tag that contains a reference to a stylesheet.

## How it works

`templ.Once` uses the `context.Context` passed to a templ component's `Render(ctx, w)` method to determine whether the content already been rendered.

The `templ.Once` function takes a single argument, which is a string that uniquely identifies the content. Use the package path of the package that contains the content as a prefix to ensure uniqueness.

If any children have already been rendered, the children are not rendered again.
