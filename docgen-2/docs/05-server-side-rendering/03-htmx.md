# HTMX

https://htmx.org can be used to selectively replace content within a web page, instead of replacing the whole page in the browser. This avoids "full-page postbacks", where the whole of the browser window is updated when a button is clicked, and results in a better user experience by reducing screen "flicker", or losing scroll position.

## Usage

Using HTMX requires:

* Installation of the HTMX client-side library.
* Modifying the HTML markup to instruct the library to perform partial screen updates.

## Installation

To install the HTMX library, download the `htmx.min.js` file and serve it via HTTP.

Then add a `<script>` tag to the `<head>` section of your HTML with the `src` attribute pointing at the file.

```html
<script src="/assets/js/htmx.min.js"></script>
```

:::info
Advanced HTMX installation and usage help is covered in the user guide at https://htmx.org.
:::

## Count example

To update the counts on the page without a full postback, the `hx-post="/"` and `hx-select="#countsForm"` attributes must be added to the `<form>` element, along with an `id` attribute to uniquely identify the element.

Adding these attributes instructs the HTMX library to replace the browser's HTTP form POST and subsequent refresh with a request from HTMX instead. HTMX issues a HTTP POST operation to the `/` endpoint, and replaces the `<form>` element with the HTML that is returned.

The `/` endpoint returns a complete HTML page instead of just the updated `<form>` element HTML. The `hx-select="#countsForm"` instructs HTMX to extract the HTML content within the `countsForm` element that is returned by the web server to replace the `<form>` element.

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
