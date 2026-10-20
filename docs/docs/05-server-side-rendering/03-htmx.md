# htmx

[htmx](https://htmx.org) can be used to selectively replace content within a web page, instead of replacing the whole page in the browser. This avoids "full-page postbacks", where the whole of the browser window is updated when a button is clicked, and results in a better user experience by reducing screen "flicker", or losing scroll position.

## Usage

Using htmx requires:

* Installation of the htmx client-side library.
* Modifying the HTML markup to instruct the library to perform partial screen updates.

## Installation

To install the htmx library, download the `htmx.min.js` file and serve it via HTTP.

Then add a `<script>` tag to the `<head>` section of your HTML with the `src` attribute pointing at the file.

```html
<script src="/assets/js/htmx.min.js"></script>
```

:::info
Advanced htmx installation and usage help is covered in the user guide at https://htmx.org.
:::

## Count example

To update the counts on the page without a full postback, the `hx-post="/"` and `hx-select="#countsForm"` attributes must be added to the `<form>` element, along with an `id` attribute to uniquely identify the element.

Adding these attributes instructs the htmx library to replace the browser's HTTP form POST and subsequent refresh with a request from htmx instead. htmx issues a HTTP POST operation to the `/` endpoint, and replaces the `<form>` element with the HTML that is returned.

The `/` endpoint returns a complete HTML page instead of just the updated `<form>` element HTML. The `hx-select="#countsForm"` instructs htmx to extract the HTML content within the `countsForm` element that is returned by the web server to replace the `<form>` element.

```templ title="components/components.templ"
templ counts(global, session int) {
	// highlight-next-line
	<form id="countsForm" action="/" method="POST" hx-post="/" hx-select="#countsForm" hx-swap="outerHTML">
		<div class="columns">
			<div class={ "column", "has-text-centered", "is-primary", border }>
				<h1 class="title is-size-1 has-text-centered">{ strconv.Itoa(global) }</h1>
				<p class="subtitle has-text-centered">Global</p>
				<div><button class="button is-primary" type="submit" name="global" value="global">+1</button></div>
			</div>
			<div class={ "column", "has-text-centered", border }>
				<h1 class="title is-size-1 has-text-centered">{ strconv.Itoa(session) }</h1>
				<p class="subtitle has-text-centered">Session</p>
				<div><button class="button is-secondary" type="submit" name="session" value="session">+1</button></div>
			</div>
		</div>
	</form>
}
```

The example can be viewed at https://d3qfg6xxljj3ky.cloudfront.net

Complete source code including AWS CDK code to set up the infrastructure is available at https://github.com/a-h/templ/tree/main/examples/counter

## Using hx-on attributes

htmx supports inline JavaScript event handlers using the `hx-on:*` attributes such as `hx-on:click`, `hx-on:submit`, etc.

Attributes starting with `on`, `hx-on`, or `data-hx-on` are treated as script attributes and expect a `templ.ComponentScript` type. This includes all of the htmx event handler syntaxes: `hx-on`, `hx-on:click`, `hx-on::after-request`, `hx-on-click`, `hx-on--after-request`, `data-hx-on:click`, and `data-hx-on-click`. Attribute names are case insensitive.

For static JavaScript, use a string literal:

```templ
<button hx-on:click="alert('Hello')">Click me</button>
```

For dynamic JavaScript with server-side data, use `templ.JSFuncCall`:

```templ
<script>
	function showMessage(msg) {
		alert(msg);
	}
</script>
<button hx-on:click={ templ.JSFuncCall("showMessage", "Hello from Go") }>Click me</button>
```

## Security

templ only sanitizes htmx event handler attributes. Other htmx attributes also execute JavaScript, and templ HTML-escapes their values, but does not sanitize them:

* `hx-vals` and `hx-headers` values that start with `js:` or `javascript:` are evaluated as JavaScript.
* `hx-vars` values are evaluated as JavaScript.
* Event filters in `hx-trigger`, e.g. `click[ctrlKey]`, are evaluated as JavaScript.

htmx has a large number of attributes and extensions, so templ does not attempt to sanitize them. Never use untrusted input in htmx attributes that execute JavaScript. To pass data to `hx-vals`, use `templ.JSONString` to serialize the value as JSON.

See the htmx [security documentation](https://htmx.org/docs/#security) for more information.
