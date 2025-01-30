# CSS style management

## HTML class and style attributes

The standard HTML `class` and `style` attributes can be added to components. Note the use of standard quotes to denote a static value.

```templ
templ button(text string) {
	<button class="button is-primary" style="background-color: red">{ text }</button>
}
```

```html title="Output"
<button class="button is-primary" style="background-color: red">
 Click me
</button>
```

## Style attribute

To use a variable in the style attribute, use braces to denote the Go expression.

```templ
templ button(style, text string) {
	<button style={ style }>{ text }</button>
}
```

You can pass multiple values to the `style` attribute. The results are all added to the output.

```templ
templ button(style1, style2 string, text string) {
	<button style={ style1, style2 }>{ text }</button>
}
```

The style attribute supports use of the following types:

* `string` - A string containing CSS properties, e.g. `background-color: red`.
* `templ.SafeCSS` - A value containing CSS properties and values that will not be sanitized, e.g. `background-color: red; text-decoration: underline`
* `map[string]string` - A map of string keys to string values, e.g. `map[string]string{"color": "red"}`
* `map[string]templ.SafeCSSProperty` - A map of string keys to values, where the values will not be sanitized.
* `templ.KeyValue[string, string]` - A single CSS key/value.
* `templ.KeyValue[string, templ.SafeCSSProperty` - A CSS key/value, but the value will not be sanitized.
* `templ.KeyValue[string, bool]` - A map where the CSS in the key is only included in the output if the boolean value is true.
* `templ.KeyValue[templ.SafeCSS, bool]` - A map where the CSS in the key is only included if the boolean value is true.

Finally, a function value that returns any of the above types can be used.

Go syntax allows you to pass a single function that returns a value and an error.

```templ
templ Page(userType string) {
	<div style={ getStyle(userType) }>Styled</div>
}

func getStyle(userType string) (string, error) {
   //TODO: Look up in something that might error.
   return "background-color: red", errors.New("failed")
}
```

Or multiple functions and values that return a single type.

```templ
templ Page(userType string) {
	<div style={ getStyle(userType), "color: blue" }>Styled</div>
}

func getStyle(userType string) (string) {
   return "background-color: red"
}
```

### Style attribute examples

#### Maps

Maps are useful when styles need to be dynamically computed based on component state or external inputs.

```templ
func getProgressStyle(percent int) map[string]string {
    return map[string]string{
        "width": fmt.Sprintf("%d%%", percent),
        "transition": "width 0.3s ease",
    }
}

templ ProgressBar(percent int) {
    <div style={ getProgressStyle(percent) } class="progress-bar">
        <div class="progress-fill"></div>
    </div>
}
```

```html title="Output (percent=75)"
<div style="transition:width 0.3s ease;width:75%;" class="progress-bar">
    <div class="progress-fill"></div>
</div>
```

#### KeyValue pattern

The `templ.KV` helper provides conditional style application in a more compact syntax.

```templ
templ TextInput(value string, hasError bool) {
    <input
        type="text"
        value={ value }
        style={
            templ.KV("border-color: #ff3860", hasError),
            templ.KV("background-color: #fff5f7", hasError),
            "padding: 0.5em 1em;",
        }
    >
}
```

```html title="Output (hasError=true)"
<input 
    type="text" 
    value="" 
    style="border-color: #ff3860; background-color: #fff5f7; padding: 0.5em 1em;">
```

#### Bypassing sanitization

By default, dynamic CSS values are sanitized to protect against dangerous CSS values that might introduce vulnerabilities into your application.

However, if you're sure, you can bypass sanitization by marking your content as safe with the `templ.SafeCSS` and `templ.SafeCSSProperty` types.

```templ
func calculatePositionStyles(x, y int) templ.SafeCSS {
    return templ.SafeCSS(fmt.Sprintf(
        "transform: translate(%dpx, %dpx);",
        x*2,  // Example calculation
        y*2,
    ))
}

templ DraggableElement(x, y int) {
    <div style={ calculatePositionStyles(x, y) }>
        Drag me
    </div>
}
```

```html title="Output (x=10, y=20)"
<div style="transform: translate(20px, 40px);">
    Drag me
</div>
```

### Pattern use cases

| Pattern | Best For | Example Use Case |
|---------|----------|------------------|
| **Maps** | Dynamic style sets requiring multiple computed values | Progress indicators, theme switching |
| **KeyValue** | Conditional style toggling | Form validation, interactive states |
| **Functions** | Complex style generation | Animations, data visualizations |
| **Direct Strings** | Simple static styles | Basic formatting, utility classes |

### Sanitization behaviour

By default, dynamic CSS values are sanitized to protect against dangerous CSS values that might introduce vulnerabilities into your application.

```templ
templ UnsafeExample() {
    <div style={ "background-image: url('javascript:alert(1)')" }>
        Dangerous content
    </div>
}
```

```html title="Output"
<div style="background-image:zTemplUnsafeCSSPropertyValue;">
    Dangerous content
</div>
```

These protections can be bypassed with the `templ.SafeCSS` and `templ.SafeCSSProperty` types.

```templ
templ SafeEmbed() {
    <div style={ templ.SafeCSS("background-image: url(/safe.png);") }>
        Trusted content
    </div>
}
```

```html title="Output"
<div style="background-image: url(/safe.png);">
    Trusted content
</div>
```

:::note
HTML attribute escaping is not bypassed, so `<`, `>`, `&` and quotes will always appear as HTML entities (`&lt;` etc.) in attributes - this is good practice, and doesn't affect how browsers use the CSS.
:::

### Error Handling

Invalid values are automatically sanitized:

```templ
templ InvalidButton() {
    <button style={ 
        map[string]string{
            "": "invalid-property",
            "color": "</style>",
        }
    }>Click me</button>
}
```

```html title="Output"
<button style="zTemplUnsafeCSSPropertyName:zTemplUnsafeCSSPropertyValue;color:zTemplUnsafeCSSPropertyValue;">
    Click me
</button>
```

Go's type system doesn't support union types, so it's not possible to limit the inputs to the style attribute to just the supported types.

As such, the attribute takes `any`, and executes type checks at runtime. Any invalid types will produce the CSS value `zTemplUnsupportedStyleAttributeValue:Invalid;`.

## Class attributes

To use a variable as the name of a CSS class, use a CSS expression.

```templ title="component.templ"
package main

templ button(text string, className string) {
	<button class={ className }>{ text }</button>
}
```

The class expression can take an array of values.

```templ title="component.templ"
package main

templ button(text string, className string) {
	<button class={ "button", className }>{ text }</button>
}
```

### Dynamic class names

Toggle addition of CSS classes to an element based on a boolean value by passing:

* A `string` containing the name of a class to apply.
* A `templ.KV` value containing the name of the class to add to the element, and a boolean that determines whether the class is added to the attribute at render time.
  * `templ.KV("is-primary", true)`
  * `templ.KV("hover:red", true)`
* A map of string class names to a boolean that determines if the class is added to the class attribute value at render time:
  * `map[string]bool`
  * `map[CSSClass]bool`

```templ title="component.templ"
package main

css red() {
	background-color: #ff0000;
}

templ button(text string, isPrimary bool) {
	<button class={ "button", templ.KV("is-primary", isPrimary), templ.KV(red(), isPrimary) }>{ text }</button>
}
```

```go title="main.go"
package main

import (
	"context"
	"os"
)

func main() {
	button("Click me", false).Render(context.Background(), os.Stdout)
}
```

```html title="Output"
<button class="button">
 Click me
</button>
```

## CSS elements

The standard `<style>` element can be used within a template.

`<style>` element contents are rendered to the output without any changes.

```templ
templ page() {
	<style type="text/css">
		p {
			font-family: sans-serif;
		}
		.button {
			background-color: black;
			foreground-color: white;
		}
	</style>
	<p>
		Paragraph contents.
	</p>
}
```

```html title="Output"
<style type="text/css">
	p {
		font-family: sans-serif;
	}
	.button {
		background-color: black;
		foreground-color: white;
	}
</style>
<p>
	Paragraph contents.
</p>
```

:::tip
If you want to make sure that the CSS element is only output once, even if you use a template many times, use a CSS expression.
:::

## CSS components

When developing a component library, it may not be desirable to require that specific CSS classes are present when the HTML is rendered.

There may be CSS class name clashes, or developers may forget to include the required CSS.

To include CSS within a component library, use a CSS component.

CSS components can also be conditionally rendered.

```templ title="component.templ"
package main

var red = "#ff0000"
var blue = "#0000ff"

css primaryClassName() {
	background-color: #ffffff;
	color: { red };
}

css className() {
	background-color: #ffffff;
	color: { blue };
}

templ button(text string, isPrimary bool) {
	<button class={ "button", className(), templ.KV(primaryClassName(), isPrimary) }>{ text }</button>
}
```

```html title="Output"
<style type="text/css">
 .className_f179{background-color:#ffffff;color:#ff0000;}
</style>
<button class="button className_f179">
 Click me
</button>
```

:::info
The CSS class is given a unique name the first time it is used, and only rendered once per HTTP request to save bandwidth.
:::

:::caution
The class name is autogenerated, don't rely on it being consistent.
:::

### CSS component arguments

CSS components can also require function arguments.

```templ title="component.templ"
package main

css loading(percent int) {
	width: { fmt.Sprintf("%d%%", percent) };
}

templ index() {
    <div class={ loading(50) }></div>
    <div class={ loading(100) }></div>
}
```

```html title="Output"
<style type="text/css">
 .loading_a3cc{width:50%;}
</style>
<div class="loading_a3cc"></div>
<style type="text/css">
 .loading_9ccc{width:100%;}
</style>
<div class="loading_9ccc"></div>
```

### CSS Sanitization

To prevent CSS injection attacks, templ automatically sanitizes dynamic CSS property names and values using the `templ.SanitizeCSS` function. Internally, this uses a lightweight fork of Google's `safehtml` package to sanitize the value.

If a property name or value has been sanitized, it will be replaced with `zTemplUnsafeCSSPropertyName` for property names, or `zTemplUnsafeCSSPropertyValue` for property values.

To bypass this sanitization, e.g. for URL values of `background-image`, you can mark the value as safe using the `templ.SafeCSSProperty` type.

```templ
css windVaneRotation(degrees float64) {
	transform: { templ.SafeCSSProperty(fmt.Sprintf("rotate(%ddeg)", int(math.Round(degrees)))) };
}

templ Rotate(degrees float64) {
	<div class={ windVaneRotation(degrees) }>Rotate</div>
}
```

### CSS Middleware

The use of CSS templates means that `<style>` elements containing the CSS are rendered on each HTTP request.

To save bandwidth, templ can provide a global stylesheet that includes the output of CSS templates instead of including `<style>` tags in each HTTP request.

To provide a global stylesheet, use templ's CSS middleware, and register templ classes on application startup.

The middleware adds a HTTP route to the web server (`/styles/templ.css` by default) that renders the `text/css` classes that would otherwise be added to `<style>` tags when components are rendered. 

For example, to stop the `className` CSS class from being added to the output, the HTTP middleware can be used.

```go
c1 := className()
handler := NewCSSMiddleware(httpRoutes, c1)
http.ListenAndServe(":8000", handler)
```

:::caution
Don't forget to add a `<link rel="stylesheet" href="/styles/templ.css">` to your HTML to include the generated CSS class names!
:::
