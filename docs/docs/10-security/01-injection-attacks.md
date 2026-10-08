# Injection attacks

templ is designed to prevent user-provided data from being used to inject vulnerabilities.

`<script>` and `<style>` tags could allow user data to inject vulnerabilities, so variables are not permitted in these sections.

```html
templ Example() {
  <script>
    function showAlert() {
      alert("hello");
    }
  </script>
  <style type="text/css">
    /* Only CSS is allowed */
  </style>
}
```

`onClick` attributes, and other `on*` attributes are used to execute JavaScript. To prevent user data from being unescaped, `on*` attributes accept a `templ.ComponentScript`. htmx `hx-on*` and `data-hx-on*` attributes are handled in the same way. Attribute names are case insensitive, so `OnClick` is handled in the same way as `onclick`.

```html
script onClickHandler(msg string) {
  alert(msg);
}

templ Example(msg string) {
  <div onClick={ onClickHandler(msg) }>
    { "will be HTML encoded using templ.Escape" }
  </div>
}
```

Style attributes cannot be expressions, only constants, to avoid escaping vulnerabilities. templ style templates (`css className()`) should be used instead.

```html
templ Example() {
  <div style={ "will throw an error" }></div>
}
```

Class names are sanitized by default. A failed class name is replaced by `--templ-css-class-safe-name`. The sanitization can be bypassed using the `templ.SafeClass` function, but the result is still subject to escaping.

```html
templ Example() {
  <div class={ "unsafe</style&gt;-will-sanitized", templ.SafeClass("&sanitization bypassed") }></div>
}
```

Rendered output:

```html
<div class="--templ-css-class-safe-name &amp;sanitization bypassed"></div>
```

```html
templ Example() {
  <div>Node text is not modified at all.</div>
  <div>{ "will be escaped using templ.EscapeString" }</div>
}
```

Attributes that contain a URL, such as `href`, `src`, `action`, and `formaction`, are sanitized with `templ.URL` to remove JavaScript URLs, unless the value is a `templ.SafeURL`. templ treats the same attributes as URLs as Go's `html/template` package. See [URL attributes](/syntax-and-usage/attributes#url-attributes) for the full list.

```html
templ Example(url string) {
  <a href="http://constants.example.com/are/not/sanitized">Text</a>
  <a href={ url }>will be sanitized by templ.URL to remove potential attacks</a>
  <a href={ templ.SafeURL("will not be sanitized by templ.URL") }</a>
}
```

The value of a `srcdoc` attribute is parsed as HTML, so a `string` value is escaped as text. Pass a templ component to render HTML.

Attributes with names that are only known at render time, i.e. [attribute key expressions](/syntax-and-usage/attributes#attribute-key-expressions) and [spread attributes](/syntax-and-usage/attributes#spread-attributes), are sanitized at render time:

* An attribute name that would change the structure of the element, e.g. `x onclick=alert(1)`, is replaced with `data-templ-failed-sanitization`.
* URL attribute values are sanitized with `templ.URL`.
* Event handler values that are not a `templ.ComponentScript` are replaced with `/* templ: failed sanitization, use templ.JSFuncCall, or templ.JSUnsafeFuncCall for trusted JavaScript */`.

templ does not know about the attributes of other JavaScript frameworks that execute their values, such as Alpine.js `x-on:*`, `@*`, `x-data`, and `x-init`, or Datastar `data-on-*`. Never use untrusted input in these attributes, or in `<script src>`.

templ only handles htmx event handler attributes. Other htmx attributes also execute JavaScript, e.g. `hx-vals` and `hx-headers` values that start with `js:` or `javascript:`, `hx-vars`, and event filters in `hx-trigger`. Never use untrusted input in these attributes.

Within css blocks, property names, and constant CSS property values are not sanitized or escaped.

```css
css className() {
	background-color: #ffffff;
}
```

CSS property values based on expressions are passed through `templ.SanitizeCSS` to replace potentially unsafe values with placeholders.

```css
css className() {
	color: { red };
}
```
