![templ](https://github.com/a-h/templ/raw/main/templ.png)

## An HTML templating language for Go that has great developer tooling.

![templ](ide-demo.gif)


## Documentation

See user documentation at https://templ.guide

<p align="center">
<a href="https://pkg.go.dev/github.com/a-h/templ"><img src="https://pkg.go.dev/badge/github.com/a-h/templ.svg" alt="Go Reference" /></a>
<a href="https://xcfile.dev"><img src="https://xcfile.dev/badge.svg" alt="xc compatible" /></a>
<a href="https://raw.githack.com/wiki/a-h/templ/coverage.html"><img src="https://github.com/a-h/templ/wiki/coverage.svg" alt="Go Coverage" /></a>
<a href="https://goreportcard.com/report/github.com/a-h/templ"><img src="https://goreportcard.com/badge/github.com/a-h/templ" alt="Go Report Card" /></a<
</p>

## Tasks

### build

Build a local version.

```sh
go run ./get-version > .version
cd cmd/templ
go build
```

### nix-update-gomod2nix

```sh
gomod2nix
```

### install-snapshot

Build and install current version.

```sh
# Remove templ from the non-standard ~/bin/templ path
# that this command previously used.
rm -f ~/bin/templ
# Clear LSP logs.
rm -f cmd/templ/lspcmd/*.txt
# Update version.
go run ./get-version > .version
# Install to $GOPATH/bin or $HOME/go/bin
cd cmd/templ && go install
```

### build-snapshot

Use goreleaser to build the command line binary using goreleaser.

```sh
goreleaser build --snapshot --clean
```

### generate

Run templ generate using local version.

```sh
go run ./cmd/templ generate -include-version=false
```

### test

Run Go tests.

```sh
go run ./get-version > .version
go run ./cmd/templ generate -include-version=false
go test ./...
```

### test-short

Run Go tests.

```sh
go run ./get-version > .version
go run ./cmd/templ generate -include-version=false
go test ./... -short
```

### test-cover

Run Go tests.

```sh
# Create test profile directories.
mkdir -p coverage/fmt
mkdir -p coverage/generate
mkdir -p coverage/version
mkdir -p coverage/unit
# Build the test binary.
go build -cover -o ./coverage/templ-cover ./cmd/templ
# Run the covered generate command.
GOCOVERDIR=coverage/fmt ./coverage/templ-cover fmt .
GOCOVERDIR=coverage/generate ./coverage/templ-cover generate -include-version=false
GOCOVERDIR=coverage/version ./coverage/templ-cover version
# Run the unit tests.
go test -cover ./... -coverpkg ./... -args -test.gocoverdir="$PWD/coverage/unit"
# Display the combined percentage.
go tool covdata percent -i=./coverage/fmt,./coverage/generate,./coverage/version,./coverage/unit
# Generate a text coverage profile for tooling to use.
go tool covdata textfmt -i=./coverage/fmt,./coverage/generate,./coverage/version,./coverage/unit -o coverage.out
# Print total
go tool cover -func coverage.out | grep total
```

### test-cover-watch

```sh
gotestsum --watch -- -coverprofile=coverage.out
```

### benchmark

Run benchmarks.

```sh
go run ./cmd/templ generate -include-version=false && go test ./... -bench=. -benchmem
```

### fmt

Format all Go and templ code.

```sh
gofmt -s -w .
go run ./cmd/templ fmt .
```

### lint

Run the lint operations that are run as part of the CI.

```sh
golangci-lint run --verbose
```

### ensure-generated

Ensure that templ files have been generated with the local version of templ, and that those files have been added to git.

Requires: generate

```sh
git diff --exit-code
```

### push-release-tag

Push a semantic version number to Github to trigger the release process.

```sh
./push-tag.sh
```

### docs-run

Run the development server.

Directory: docs

```sh
npm run start
```

### docs-build

Build production docs site.

Directory: docs

```sh
npm run build
```

