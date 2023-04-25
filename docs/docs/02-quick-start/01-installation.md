# Installation

## go install

With Go 1.20 or greater installed, run:

```sh
go install github.com/a-h/templ/cmd/templ@latest
```

## Github binaries

Download the latest release from https://github.com/a-h/templ/releases/latest

## Nix

templ provides a Nix flake with an exported package containing the binary at https://github.com/a-h/templ/blob/main/flake.nix

```sh
nix run github:a-h/templ
```

templ also provides a development shell which includes a Neovim configuration setup to use the templ autocompletion features.

```sh
nix develop github:a-h/templ
```
