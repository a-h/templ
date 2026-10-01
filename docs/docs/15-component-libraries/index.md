# Component Libraries

Component libraries in the templ ecosystem provide ready-to-use UI elements.

## shadcn-templ

![shadcn-templ Banner](/img/ecosystem/shadcn-templ.png)

### About

shadcn-templ (formerly templUI) is a 1:1 port of shadcn/ui for Go and templ. Same API surface, markup and styles as the original, with the Base UI client behavior ported to dependency-free vanilla JavaScript. The CLI copies component source into your project, so the code is yours.

### Features

- **50+ components**, rebuilt 1:1 against shadcn/ui
- **Eight visual styles** and a theme builder
- **No npm, no Node**: server-rendered, dependency-free vanilla JS
- **CLI with project templates, presets and a registry** that compiles every component for your chosen style

### Usage

```shell
go install github.com/axadrn/shadcn-templ/v2/cmd/shadcn-templ@latest
shadcn-templ init my-app --template templ
```

### Links

- [Documentation](https://shadcn-templ.com)
- [GitHub](https://github.com/axadrn/shadcn-templ)

## templ-components

### About

templ-components is a server-rendered UI component library built on templ, HTMX, and Tailwind CSS v4. It ships typed Go props, dark mode, CSP nonce support, and ARIA accessibility out of the box, with no client-side JavaScript framework — JavaScript only enhances the server-rendered HTML.

### Features

- **120+ typed components**: Cards, tables, forms, overlays, charts, and a dedicated HTMX package for loading, error, and out-of-band swap helpers
- **Dual-transport wiring**: One typed action spec renders either HTMX attributes or Datastar expressions
- **Dark mode and accessibility**: Every component ships tested dark-mode variants, ARIA roles, and keyboard navigation
- **No framework lock-in**: Pure Go with templ and Tailwind v4 class strings; no Node.js runtime
- **CSP-safe by construction**: Every inline script carries a nonce

### Example

```go
import (
  "github.com/larsartmann/templ-components/display"
  "github.com/larsartmann/templ-components/layout"
)

templ ExamplePage() {
  @layout.Base(layout.DefaultPageProps()) {
    @display.StatCard(display.StatCardProps{
      Label:  "Active users",
      Value:  "1280",
      Change: "+12%",
      Trend:  display.TrendUp,
    })
  }
}
```

### Links

- [Documentation](https://templcomponents.lars.software)
- [GitHub](https://github.com/larsartmann/templ-components)
- [API Reference](https://pkg.go.dev/github.com/larsartmann/templ-components)
