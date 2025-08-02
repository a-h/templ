# Component Libraries

Component libraries in the templ ecosystem provide ready-to-use UI elements.

## templUI

![templUI Banner](/img/ecosystem/templui.png)

### About

templUI is the premier UI component library built specifically for templ. It combines the type-safety of Go with the interactivity of Alpine.js and the styling power of Tailwind CSS to create beautiful, responsive web applications.

### Features

- **30+ Ready-made Components**: Buttons, cards, modals, charts, and more
- **Enterprise-Ready**: Built for production with security in mind
- **CSP Compliant**: Works seamlessly with Content Security Policy
- **Type-Safe**: Full Go type system integration and checking
- **Customizable**: Easily adapt to match your brand identity

### Example

```go
import "github.com/axzilla/templui/components"

templ ExamplePage() {
  @components.Button(components.ButtonProps{
    Text: "Click me",
    IconRight: icons.ArrowRight(icons.IconProps{Size: "16"}),
  })
}
```

### Links

- [Documentation](https://templui.io)
- [GitHub](https://github.com/axzilla/templui)
- [Quick Start Template](https://github.com/axzilla/templui-quickstart)

## DatastarUI

### About

DatastarUI is a comprehensive UI component library that ports shadcn/ui components to Go and templ. It provides pixel-perfect visual and functional replicas of popular UI components with reactive capabilities powered by Datastar signals.

### Features

- **shadcn/ui Port**: Faithful recreation of shadcn/ui components in Go/templ
- **Reactive UI**: Powered by lightweight Datastar signals (15KB runtime)
- **Server-side Rendered**: All components render on the server for optimal performance
- **Type-safe**: Full Go type system integration with structured component arguments
- **Tailwind CSS**: Consistent styling that matches shadcn/ui design system
- **Dark Mode**: Built-in dark mode support across all components
- **Accessibility**: Focused on creating accessible UI components

### Example

```go
import "github.com/CoreyCole/datastarui/components"

templ ExamplePage() {
  @components.Button(components.ButtonProps{
    ID: "my-button",
    Variant: "primary",
    Size: "md",
    Loading: false,
  }) {
    Click me
  }
}
```

### Development Setup

DatastarUI includes a streamlined development workflow:

```sh
# Start Tailwind CSS watcher and Go server with live reload
templ generate --watch --proxy="http://localhost:4242" --cmd="go run ."
```

### Links

- [GitHub](https://github.com/CoreyCole/datastarui)
- [Datastar](https://data-star.dev/) - The reactive framework powering the components
