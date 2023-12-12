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

To install in your Nix Flake:

This flake exposes an overlay, so you can add it to your own Flake and/or NixOS system.

```nix
{
  inputs = {
    ...
    templ.url = "github:a-h/templ";
    ...
  };
  outputs = inputs@{
    ...
  }:

  # For NixOS configuration:
  {
    # Add the overlay,
    nixpkgs.overlays = [
      inputs.templ.overlays.default
    ];
    # and install the package
    environment.systemPackages = with pkgs; [
      templ
    ];
  };

  # For a flake project:
  let
    forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
      inherit system;
      pkgs = import nixpkgs { inherit system; };
    });
    templ = system: inputs.templ.packages.${system}.templ;
  in {
    packages = forAllSystems ({ pkgs, system }): {
      myNewPackage = pkgs.buildGoModule {
        ...
        preBuild = ''
          ${templ system}/bin/templ generate
        '';
      };
    };

    devShell = forAllSystems ({ pkgs, system }):
      pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          (templ system)
        ];
      };
  };
}
```
