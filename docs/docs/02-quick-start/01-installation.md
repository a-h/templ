# Installation

## go install

With Go 1.23 or greater installed, run:

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

## Github binaries

Download the latest release from https://github.com/a-h/templ/releases/latest

## Nix

templ provides a Nix flake with an exported package containing the binary at https://github.com/a-h/templ/blob/main/flake.nix

```bash
nix run github:a-h/templ
```

templ also provides a development shell which includes all of the tools required to build templ, e.g. go, gopls etc. but not templ itself.

```bash
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
    packages = forAllSystems ({ pkgs, system }: {
      myNewPackage = pkgs.buildGoModule {
        ...
        preBuild = ''
          ${templ system}/bin/templ generate
        '';
      };
    });

    devShell = forAllSystems ({ pkgs, system }:
      pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          (templ system)
        ];
      };
  });
}
```

## Docker

A Docker container is pushed on each release to https://github.com/a-h/templ/pkgs/container/templ

Pull the latest version with:

```bash
docker pull ghcr.io/a-h/templ:latest
```

To use the container, mount the source code of your application into the `/app` directory, set the working directory to the same directory and run `templ generate`, e.g. in a Linux or Mac shell, you can generate code for the current directory with:

```bash
docker run -v `pwd`:/app -w=/app ghcr.io/a-h/templ:latest generate
```

If you want to build templates using a multi-stage Docker build, you can use the `templ` image as a base image.

Here's an example multi-stage Dockerfile. Note that in the `generate-stage` the source code is copied into the container, and the `templ generate` command is run. The `build-stage` then copies the generated code into the container and builds the application.

The permissions of the source code are set to a user with a UID of 65532, which is the UID of the `nonroot` user in the `ghcr.io/a-h/templ:latest` image.

Note also the use of the `RUN ["templ", "generate"]` command instead of the common `RUN templ generate` command. This is because the templ Docker container does not contain a shell environment to keep its size minimal, so the command must be ran in the ["exec" form](https://docs.docker.com/reference/dockerfile/#shell-and-exec-form).

```Dockerfile
# Fetch
FROM golang:latest AS fetch-stage
COPY go.mod go.sum /app
WORKDIR /app
RUN go mod download

# Generate
FROM ghcr.io/a-h/templ:latest AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

# Build
FROM golang:latest AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app

# Test
FROM build-stage AS test-stage
RUN go test -v ./...

# Deploy
FROM gcr.io/distroless/base-debian12 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```
