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
