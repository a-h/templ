# Attributes

## Constant attributes

templ elements can have HTML attributes that use the double quote character `"`.

```templ
templ component() {
  <p data-testid="paragraph">Text</p>
}
```

```html title="Output"
<p data-testid="paragraph">Text</p>
```

## String expression attributes

Element attributes can be set to Go strings.

```templ
templ component(testID string) {
  <p data-testid={ testID }>Text</p>
}

templ page() {
  @component("testid-123")
}
```

Rendering the `page` component results in:

```html title="Output"
<p data-testid="testid-123">Text</p>
```

:::note
String values are automatically HTML attribute encoded. This is a security measure, but may make the values (especially JSON appear) look strange to you, since some characters may be converted into HTML entities. However, it is correct HTML and won't affect the behavior. 
:::

It's also possible to use function calls in string attribute expressions.

Here's a function that returns a string based on a boolean input.

```go
func testID(isTrue bool) string {
    if isTrue {
        return "testid-123"
    }
    return "testid-456"
}
```

```templ
templ component() {
  <p data-testid={ testID(true) }>Text</p>
}
```

The result:

```html title="Output"
<p data-testid="testid-123">Text</p>
```

Functions in string attribute expressions can also return errors.

```go
func testID(isTrue bool) (string, error) {
    if isTrue {
        return "testid-123", nil
    }
    return "", fmt.Errorf("isTrue is false")
}
```

If the function returns an error, the `Render` method will return the error along with its location.

## Boolean attributes

Boolean attributes (see https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#boolean-attributes) where the presence of an attribute name without a value means true, and the attribute name not being present means false are supported.

```templ
templ component() {
  <hr noshade/>
}
```

```html title="Output"
<hr noshade>
```

:::note
templ is aware that `<hr/>` is a void element, and renders `<hr>` instead.
:::


To set boolean attributes using variables or template parameters, a question mark after the attribute name is used to denote that the attribute is boolean.

```templ
templ component() {
  <hr noshade?={ false } />
}
```

```html title="Output"
<hr>
```

## Conditional attributes

Use an `if` statement within a templ element to optionally add attributes to elements.

```templ
templ component() {
  <hr style="padding: 10px"
    if true {
      class="itIsTrue"
    }
  />
}
```

```html title="Output"
<hr style="padding: 10px" class="itIsTrue" />
```

## Attribute key expressions

Use a string expression to dynamically set the key of an attribute.

```templ
templ paragraph(testID string) {
  <p { "data-" + testID }="paragraph">Text</p>
}

templ component() {
  @paragraph("testid")
}
```

```html title="Output"
<p data-testid="paragraph">Text</p>
```

The name of the attribute is not known until the template is rendered, so templ sanitizes the attribute at render time:

* If the name would change the structure of the element, e.g. because it contains a space, a quote, or an `=` sign, templ replaces the name with `data-templ-failed-sanitization`.
* If the attribute is a [URL attribute](#url-attributes), templ sanitizes the value as a URL.
* If the attribute is an [event handler](#javascript-attributes), and the value is not a `templ.ComponentScript`, templ replaces the value with `/* templ: failed sanitization, use templ.JSFuncCall, or templ.JSUnsafeFuncCall for trusted JavaScript */`.
* If the attribute is a `srcdoc` attribute, templ escapes a `string` value as text.
* The `&` character can't be used in the name, because browsers don't decode character references in attribute names.

```templ
templ component(name string, value string) {
  <a { name }={ value }>Text</a>
}

templ usage() {
  @component("href", "javascript:alert(1)")
  @component("onclick", "alert(1)")
  @component("x onmouseover=alert(1) y", "value")
}
```

```html title="Output"
<a href="about:invalid#TemplFailedSanitizationURL">Text</a>
<a onclick="/* templ: failed sanitization, use templ.JSFuncCall, or templ.JSUnsafeFuncCall for trusted JavaScript */">Text</a>
<a data-templ-failed-sanitization="value">Text</a>
```

To set an event handler with an attribute key expression, pass a `templ.ComponentScript` as the value:

* `templ.JSFuncCall("functionName", args...)` calls a JavaScript function, and JSON encodes the Go arguments, so it's safe to use with untrusted data. See [Pass Go data to a JavaScript event handler](/syntax-and-usage/script-templates#pass-go-data-to-a-javascript-event-handler).
* `templ.JSUnsafeFuncCall("alert('hello')")` renders the string as JavaScript without escaping. Only use it for JavaScript that comes from a trusted source, never with user input. See [Call client side functions with server side data](/syntax-and-usage/script-templates#call-client-side-functions-with-server-side-data).

```templ
templ component(name string, message string) {
  <button { name }={ templ.JSFuncCall("alert", message) }>Show message</button>
  <button { name }={ templ.JSUnsafeFuncCall("alert('hello')") }>Say hello</button>
}

templ usage() {
  @component("onclick", "Hello, World!")
}
```

```html title="Output"
<button onclick="alert(&#34;Hello, World!&#34;)">Show message</button>
<button onclick="alert(&#39;hello&#39;)">Say hello</button>
```

## Spread attributes

Use the `{ attrMap... }` syntax in the open tag of an element to append a dynamic map of attributes to the element's attributes.

It's possible to spread any variable of type `templ.Attributes`. `templ.Attributes` is a `map[string]any` type definition.

* If the value is a `string`, the attribute is added with the string value, e.g. `<div name="value">`.
* If the value is a `bool`, the attribute is added as a boolean attribute if the value is true, e.g. `<div name>`.
* If the value is a `templ.KeyValue[string, bool]`, the attribute is added if the boolean is true, e.g. `<div name="value">`.
* If the value is a `templ.KeyValue[bool, bool]`, the attribute is added if both boolean values are true, as `<div name>`.
* If the value is a `templ.SafeURL`, the attribute is added with the URL value, without URL sanitization.
* If the value is a `templ.ComponentScript`, i.e. the result of `templ.JSFuncCall` or `templ.JSUnsafeFuncCall`, the attribute is added with the JavaScript as the value.

Spread attributes are sanitized in the same way as [attribute key expressions](#attribute-key-expressions), so `templ.Attributes{"href": url}` is sanitized as a URL, and `templ.Attributes{"onclick": "alert(1)"}` is replaced, because the value is a `string`, not a `templ.ComponentScript`. To set an event handler, use `templ.JSFuncCall` to call a JavaScript function with Go data, e.g. `templ.Attributes{"onclick": templ.JSFuncCall("alert", message)}`, or `templ.JSUnsafeFuncCall` for JavaScript from a trusted source, e.g. `templ.Attributes{"onclick": templ.JSUnsafeFuncCall("alert('hello')")}`.

```templ
templ component(shouldBeUsed bool, attrs templ.Attributes) {
  <p { attrs... }>Text</p>
  <hr
    if shouldBeUsed {
      { attrs... }
    }
  />
}

templ usage() {
  @component(false, templ.Attributes{"data-testid": "paragraph"}) 
}
```

```html title="Output"
<p data-testid="paragraph">Text</p>
<hr>
```

## URL attributes

Attributes that expect a URL, such as `<a href={ url }>`, `<form action={ url }>`, or `<img src={ url }>`, have special behavior if you use a dynamic value.

```templ
templ component(p Person) {
  <a href={ p.URL }>{ strings.ToUpper(p.Name) }</a>
}
```

When you pass a `string` to these attributes, templ will automatically sanitize the input URL, ensuring that the protocol is safe (e.g., `http`, `https`, or `mailto`) and does not contain potentially harmful protocols like `javascript:`. A URL that fails sanitization is replaced with `about:invalid#TemplFailedSanitizationURL`.

templ treats the same attributes as URLs as Go's `html/template` package does, on any element:

* `action`, `archive`, `background`, `cite`, `classid`, `codebase`, `data`, `formaction`, `href`, `icon`, `longdesc`, `manifest`, `poster`, `profile`, `src`, `usemap`, and `xmlns`.
* Namespaced attributes with one of the names above, e.g. `xlink:href`, and all `xmlns:*` attributes.
* `data-*` attributes with one of the names above, e.g. `data-href`.
* Any attribute with a name that contains `src`, `uri`, or `url`, e.g. `data-url`, because developers often store URLs in custom attributes, and JavaScript may navigate to them.

`srcset` contains `src`, so templ sanitizes it as a single URL. `html/template` sanitizes each URL in a `srcset` separately.

templ also sanitizes `to`, `from`, `values`, and `by` on SVG `<set>` and `<animate>` elements, which can set the `href` of their parent element, unless the `attributeName` attribute is a constant that can't refer to `href`, i.e. it isn't `href`, or `href` with a namespace prefix such as `xlink:href`, ignoring case and surrounding whitespace.

Attribute names are case insensitive, so `<a HREF={ url }>` is sanitized in the same way as `<a href={ url }>`.

`data:` and `blob:` URLs fail sanitization, including in `<img src>`. To use a `data:` URL, e.g. for an inline image, convert it to a `templ.SafeURL`.

```templ
templ component(url string) {
  <a HREF={ url }>Link</a>
  <iframe src={ url }></iframe>
  <img src={ "data:image/gif;base64,R0lGODlhAQABAAAAACw=" }/>
  <img src={ templ.SafeURL("data:image/gif;base64,R0lGODlhAQABAAAAACw=") }/>
}

templ usage() {
  @component("javascript:alert(1)")
}
```

```html title="Output"
<a HREF="about:invalid#TemplFailedSanitizationURL">Link</a>
<iframe src="about:invalid#TemplFailedSanitizationURL"></iframe>
<img src="about:invalid#TemplFailedSanitizationURL">
<img src="data:image/gif;base64,R0lGODlhAQABAAAAACw=">
```

:::warning
Sanitization only checks the URL scheme. A script from any `https:` URL executes in the page, so never use untrusted input for the source of a script, e.g. `<script src={ url }>`.
:::

:::caution
To bypass URL sanitization, you can use `templ.SafeURL(myURL)` to mark that your string is safe to use.

This may introduce security vulnerabilities to your program.
:::

If you use a constant value, e.g. `<a href="javascript:alert('hello')">`, templ will not modify it, and it will be rendered as is.

:::tip
Non-standard HTML attributes can contain URLs, for example htmx's `hx-*` attributes).

To sanitize URLs in that context, use the `templ.URL(urlString)` function.

```templ
templ component(contact model.Contact) {
  <div hx-get={ templ.URL(fmt.Sprintf("/contacts/%s/email", contact.ID)) }>
    { contact.Name }
  </div>
}
```
:::

:::note
In templ, all attributes are HTML-escaped. This means that:

- `&` characters in the URL are escaped to `&amp;`.
- `"` characters are escaped to `&quot;`.
- `'` characters are escaped to `&#39;`.

This done to prevent XSS attacks. For example, without escaping, if a string contained `http://google.com" onclick="alert('hello')"`, the browser would interpret this as a URL followed by an `onclick` attribute, which would execute JavaScript code.

The escaping does not change the URL's functionality.

Sanitization is the process of examining the URL scheme (protocol) and structure to ensure that it's safe to use, e.g. that it doesn't contain `javascript:` or other potentially harmful schemes. If a URL is not safe, templ will replace the URL with `about:invalid#TemplFailedSanitizationURL`.
:::

## JavaScript attributes

`onClick` and other `on*` handlers have special behaviour, they expect a `templ.ComponentScript`, such as a reference to a `script` template, or the result of `templ.JSFuncCall`.

Attribute names are case insensitive, so `OnClick` and `ONCLICK` are handled in the same way as `onclick`. htmx event handler attributes are handled in the same way: `hx-on`, `hx-on:*`, `hx-on-*`, `data-hx-on:*`, and `data-hx-on-*`.

:::warning
templ does not know about the attributes of other JavaScript frameworks that execute their values, such as Alpine.js `x-on:*`, `@*`, `x-data`, and `x-init`, or Datastar `data-on-*`. templ HTML-escapes the values of these attributes, but does not prevent them from executing JavaScript, so never use untrusted input in them.

The same applies to htmx. templ requires a `templ.ComponentScript` in htmx event handler attributes, but other htmx attributes also execute JavaScript, e.g. `hx-vals` and `hx-headers` values that start with `js:` or `javascript:`, `hx-vars`, and event filters in `hx-trigger`, such as `click[ctrlKey]`. templ does not sanitize these attributes, so never use untrusted input in them.
:::

:::info
This ensures that any client-side JavaScript that is required for a component to function is only emitted once, that script name collisions are not possible, and that script input parameters are properly sanitized.
:::

```templ
script withParameters(a string, b string, c int) {
	console.log(a, b, c);
}

script withoutParameters() {
	alert("hello");
}

templ Button(text string) {
	<button onClick={ withParameters("test", text, 123) } onMouseover={ withoutParameters() } type="button">{ text }</button>
}
```

```html title="Output"
<script>
 function __templ_withParameters_1056(a, b, c){console.log(a, b, c);}function __templ_withoutParameters_6bbf(){alert("hello");}
</script>
<button onclick="__templ_withParameters_1056("test","Say hello",123)" onmouseover="__templ_withoutParameters_6bbf()" type="button">
 Say hello
</button>
```

## iframe srcdoc attributes

The browser parses the value of an `<iframe>` `srcdoc` attribute as an HTML document, in the context of the page. templ escapes a `string` value as text, so that it is displayed, and not parsed as HTML.

To render HTML in a `srcdoc` attribute, pass a templ component. templ renders the component, and escapes the output for use in the attribute.

```templ
templ preview(text string) {
  <iframe srcdoc={ text }></iframe>
  <iframe srcdoc={ previewContent(text) }></iframe>
}

templ previewContent(text string) {
  <p>{ text }</p>
}

templ usage() {
  @preview("<b>Hello</b>")
}
```

The first `<iframe>` displays the text `<b>Hello</b>`. The second displays a paragraph that contains the same text, because `previewContent` escapes `text`.

```html title="Output"
<iframe srcdoc="&amp;lt;b&amp;gt;Hello&amp;lt;/b&amp;gt;"></iframe>
<iframe srcdoc="&lt;p&gt;&amp;lt;b&amp;gt;Hello&amp;lt;/b&amp;gt;&lt;/p&gt;"></iframe>
```

## CSS attributes

CSS handling is discussed in detail in [CSS style management](/syntax-and-usage/css-style-management).

## JSON attributes

To set an attribute's value to a JSON string (e.g. for htmx's [hx-vals](https://htmx.org/attributes/hx-vals) or Alpine's [x-data](https://alpinejs.dev/directives/data)), use `templ.JSONString` to serialize the value.

```templ
templ SearchBox(countries []string) {
	<search-webcomponent suggestions={ templ.JSONString(countries) } />
}
```
