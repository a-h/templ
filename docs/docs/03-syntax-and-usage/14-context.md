# Context

## What problems does `context` solve?

### "Prop drilling"

It can be cumbersome to pass data from parents through to children components, since this means that every component in the hierarchy has to accept parameters and pass them through to children.

The technique of passing data through a stack of components is sometimes called "prop drilling".

In this example, the `middle` component doesn't use the `name` parameter, but must accept it as a parameter in order to pass it to the `bottom` component.

```templ title="component.templ"
package main

templ top(name string) {
	<div>
		@middle(name)
	</div>
}

templ middle(name string) {
	<ul>
		@bottom(name)
	</ul>
}

templ bottom(name string) {
  <li>{ name }</li>
}
```

:::tip
In many cases, prop drilling is the best way to pass data because it's simple and reliable.

Context is not strongly typed, and errors only show at runtime, not compile time, so it should be used sparingly in your application.
:::

### Coupling

Some data is useful for many components throughout the hierarchy, for example:

* Whether the current user is logged in or not.
* The username of the current user.
* The locale of the user (used for localization).
* Theme preferences (e.g. light vs dark).

One way to pass this information is to create a `Settings` struct and pass it through the stack as a parameter.

```templ title="component.templ"
package main

type Settings struct {
	Username string
	Locale   string
	Theme    string
}

templ top(settings Settings) {
	<div>
		@middle(settings)
	</div>
}

templ middle(settings Settings) {
	<ul>
		@bottom(settings)
	</ul>
}

templ bottom(settings Settings) {
  <li>{ settings.Theme }</li>
}
```

However, this `Settings` struct may be unique to a single website, and reduce the ability to reuse a component in another website, due to its tight coupling with the `Settings` struct.

## Using `context`

:::info
templ components have an implicit `ctx` variable within the scope. This `ctx` variable is the variable that is passed to the `templ.Component`'s `Render` method.
:::

To allow data to be accessible at any level in the hierarchy, we can use Go's built in `context` package.

Within templ components, use the implicit `ctx` variable to access the context.

```templ title="component.templ"
templ themeName() {
	<div>{ ctx.Value(themeContextKey).(string) }</div>
}
```

To allow the template to get the `themeContextKey` from the context, create a context, and pass it to the component's `Render` function.

```templ title="main.go"
// Define the context key type.
type contextKey string

// Create a context key for the theme.
var themeContextKey contextKey = "theme"

// Create a context variable that inherits from a parent, and sets the value "test".
ctx := context.WithValue(context.Background(), themeContextKey, "test")

// Pass the ctx variable to the render function.
themeName().Render(ctx, w)
```

:::warning
Attempting to access a context key that doesn't exist, or using an invalid type assertion will trigger a panic.
:::

### Tidying up

Rather than read from the context object directly, it's common to implement a type-safe function instead.

This is also required when the type of the context key is in a different package to the consumer of the context, and the type is private (which is usually the case).

```templ title="main.go"
func GetTheme(ctx context.Context) string {
	if theme, ok := ctx.Value(themeContextKey).(string); ok {
		return theme
	}
	return ""
}
```

This minor change makes the template code a little tidier.

```templ title="component.templ"
templ themeName() {
	<div>{ GetTheme(ctx) }</div>
}
```

:::note
As of v0.2.731, Go's built in `context` package is no longer implicitly imported into .templ files.
:::

## Using `context` with HTTP middleware

In HTTP applications, a common pattern is to insert HTTP middleware into the request/response chain.

Middleware can be used to update the context that is passed to other components. Common use cases for middleware include authentication, and theming.

By inserting HTTP middleware, you can set values in the context that can be read by any templ component in the stack for the duration of that HTTP request.

```templ title="component.templ"
type contextKey string
var contextClass = contextKey("class")

func Middleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request ) {
    ctx := context.WithValue(r.Context(), contextClass, "red")
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

templ Page() {
  @Show()
}

templ Show() {
  <div class={ ctx.Value(contextClass) }>Display</div>
}

func main() {
  h := templ.Handler(Page())
  withMiddleware := Middleware(h)
  http.Handle("/", withMiddleware)
  http.ListenAndServe(":8080", h)
}
```

:::warning
If you write a component that relies on a context variable that doesn't exist, or is an unexpected type, your component will panic at runtime.

This means that if your component relies on HTTP middleware that sets the context, and you forget to add it, your component will panic at runtime.
:::

