# HTML Element Components

HTML Element Components enable component authoring using HTML element syntax, providing a more familiar way to invoke templ components.

## Basic Syntax

HTML Element Components use Title case naming and can be invoked using HTML-like syntax:

```go
templ Button(title string) {
  <button>{ title }</button>
}

templ Page() {
  <Button title="Click me" />
}
```

## Component with Children

Components can accept children using the `children...` placeholder:

```go
templ Card(title string) {
  <article class="card">
    <h2>{ title }</h2>
    <main>
      { children... }
    </main>
  </article>
}

templ Page() {
  <Card title="Card Title">
    <p>This is the card content.</p>
  </Card>
}
```

## Attribute Types

HTML Element Components support various attribute types:

### String Attributes

```go
// Constant string
<Button title="Constant String" />

// Expression
<Button title={ "Hello " + name + "!" } />

// Expression with error handling, func strWithError() returns (string, error)
<Button title={ strWithError() } />
```

Other primitive types like `int`, `float`, and `bool` can also be used directly as attributes, as long as the types match:

```go
templ NumberComponent(count int) {
  <div>Count: { count }</div>
}
// Integer attribute
<NumberComponent count={ 42 } />
<NumberComponent count={ int(42), error(nil) } />
```

### Boolean Attributes

```go
templ BoolComponent(enabled bool) {
  <div>
    if enabled {
      <span>Enabled</span>
    }
  </div>
}

// Boolean const attribute
<BoolComponent enabled />

// Boolean expression with ?=
<BoolComponent enabled?={ shouldEnable() } />
```

### Multiple Parameter Types

```go
templ MultiComponent(title string, count int, enabled bool) {
  <div>
    <h3>{ title } (count: { fmt.Sprint(count) })</h3>
    if enabled {
      <span>Enabled</span>
    }
  </div>
}

<MultiComponent title="Multi Test" count={ 42 } enabled />
```

:::note
HTML element component attributes are mapped to function parameters by name. The attribute `title` corresponds to the parameter `title`, `count` to `count`, and so on. Parameter order in the function signature doesn't matter - attributes are matched by name.
:::

## Inline Component Attributes

Components can accept other components as attributes using inline syntax:

```go
templ Container(child templ.Component) {
  <div class="container">
    @child
  </div>
}

// Inline component as attribute
<Container child={ <span>I love templ</span> } />

// Multi-line inline component
<Container
  child={
    <div>
      <p>Complex content</p>
    </div>
  }
/>
```

### Alternative Syntax for Component Parameters

For component parameters of type `templ.Component`, you can use curly brace syntax `{ paramName }` as an alternative to the `@paramName` syntax:

```go
templ Container(child templ.Component) {
  <div class="container">
    @child          // Traditional syntax
    { child }       // Alternative syntax - works the same
  </div>
}
```

:::note
The curly brace syntax `{ paramName }` only works for **local template parameters** of type `templ.Component`. It does not work for global variables or other types. For global component variables, you must still use the `@variableName` syntax.
:::

### Primitive Types in Inline Components

Primitive types can be passed directly to inline component attributes:

```go
templ Container(child templ.Component) {
  <div class="container">
    @child
  </div>
}

// String as inline component
<Container child="hello" />

// Expression returning string with error
<Container child={ strErr() } />
```

## Types that Implement `templ.Component`

Types that implement the `templ.Component` interface can be used directly as HTML element components:

```go
import "github.com/example/mod"

type StructComponent struct {
  Name  string
  Child templ.Component
  Attrs templ.Attributer
}

func (c *StructComponent) Render(ctx context.Context, w io.Writer) error {
  if _, err := fmt.Fprint(w, "<div class=\"struct-component\""); err != nil {
    return err
  }
  if c.Attrs != nil {
    if err := templ.RenderAttributes(ctx, w, c.Attrs); err != nil {
      return err
    }
  }
  if _, err := fmt.Fprint(w, ">"); err != nil {
    return err
  }
  if c.Child != nil {
    if err := c.Child.Render(ctx, w); err != nil {
      return err
    }
  }
  if _, err := fmt.Fprint(w, c.Name); err != nil {
    return err
  }
  if _, err := fmt.Fprint(w, "</div>"); err != nil {
    return err
  }
  return nil
}

// Usage as HTML element component
<mod.StructComponent
  Name="struct"
  Child={
    <span>struct component child</span>
  }
  enabled?={ isEnabled() }
/>
```

:::note
For structs that implement `templ.Component`, HTML element component attributes are mapped to struct fields by name. The attribute `Name` corresponds to the struct field `Name`, `Child` to `Child`, etc. Field names must be exported (capitalized) to be accessible.
:::

## External Package Components

Components from external packages are supported:

```go
import extern "github.com/example/components"

templ Page() {
  <extern.Button title="External Button" />
}
```

## Struct Components

Components can be methods on structs:

```go
type StructComponent struct{}

templ (StructComponent) Page(title string, attrs templ.Attributer) {
  <div { attrs... }>
    <h1>{ title }</h1>
  </div>
}

var structComp StructComponent

templ Usage() {
  <structComp.Page title="Struct Component" class="example" />
}
```

:::note
For struct method components, attributes are mapped to method parameters by name, just like regular templ components. The struct variable name (like `structComp`) is used to access the method.
:::

## Rest Attributes

Components can accept additional attributes using `templ.Attributer`:

```go
templ FlexibleComponent(title string, attrs templ.Attributer) {
  <div { attrs... }>
    <h1>{ title }</h1>
  </div>
}

<FlexibleComponent
  title="Flexible"
  class="custom-class"
  style="color: blue;"
  data-value="123"
/>
```

## Conditional Attributes

HTML attributes can be conditionally applied:

```go
<Button
  title="Conditional"
  if enabled {
    class="enabled"
  } else {
    class="disabled"
  }
/>
```

## Multiple Attributes

Multiple CSS properties and attributes can be combined:

```go
<Component
  style={ "color: green;", "font-weight: bold;" }
  x-data-custom="custom-value"
  disabled?={ false }
/>
```

## Function Call Attributes

JavaScript function calls can be used as attribute values:

```go
<Button onclick={ templ.JSFuncCall("alert", templ.JSExpression("hello")) } />
```

## No Arguments Components

Components without parameters can be used as self-closing elements:

```go
templ NoArgsComponent() {
  <div class="simple"></div>
}

<NoArgsComponent />
```
