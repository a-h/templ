# Live reload with other tools

Browser live reload allows you to see your changes immediately without having to switch to your browser and press `F5` or `CMD+R`.

However, Web projects usually involve multiple build processes, e.g. css bundling, js bundling, alongside templ code generation and Go compilation.

Tools like `air` can be used with templ's built-in proxy server to carry out additional steps.

## Example

This example, demonstrates setting up a live reload environment that integrates:

- [Tailwind CSS](https://tailwindcss.com/) for generating a css bundle.
- [esbuild](https://esbuild.github.io/) for bundling JavaScript or TypeScript.
- [air](https://github.com/cosmtrek/air) for re-building Go source as well as sending a reload event to the `templ` proxy server.

## How does it work

templ's built-in proxy server automatically refreshes the browser when a file changes. The proxy server injects a script that reloads the page in the browser if a "reload" event is sent to the browser by the proxy. See [Live Reload page](/developer-tools/live-reload) for a detailed explanation.

:::tip
The live reload JavaScript is only injected by the templ proxy if your HTML file contains a closing `</body>` tag.
:::

The "reload" event can be triggered in two ways:

- `templ generate --watch` sends the event whenever a ".templ" file changes.
- Manually trigger it by sending a HTTP POST request to `/_templ/reload/event` endpoint. The `templ` CLI provides this via `templ generate --notify-proxy`.

:::tip
templ proxy server `--watch` mode generates different `_templ.go` files. In `--watch` mode `_templ.txt` files are generated that contain just the text that's in templ files. This is used to skip compilation of the Go code when only the text content changes.
:::

## Setting up the Makefile

A `Makefile` can be used to run all of the necessary commands in parallel. This is useful for starting all of the watch processes at once.

### templ watch mode

To start the `templ` proxy server in watch mode, run:

```bash
templ generate --watch --proxy="http://localhost:8080" --open-browser=false
```

This assumes that your http server is running on `http://localhost:8080`. `--open-browser=false` is to prevent `templ` from opening the browser automatically.

### Tailwind CSS

Tailwind requires a `tailwind.config.js` file at the root of your project, alongside an `input.css` file.

```bash
npx --yes tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch
```

This will watch `input.css` as well as your `.templ` files and re-generate `assets/styles.css` whenever there's a change.

### esbuild

To bundle JavaScript, TypeScript, JSX, or TSX files, you can use `esbuild`:

```bash
npx --yes esbuild js/index.ts --bundle --outdir=assets/ --watch
```

This will watch `js/index.ts` and relevant files, and re-generate `assets/index.js` whenever there's a change.

### Re-build Go source

To watch and restart your Go server, when only the `go` files change you can use `air`:

```bash
go run github.com/cosmtrek/air@v1.51.0 \
  --build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
  --build.exclude_dir "node_modules" \
  --build.include_ext "go" \
  --build.stop_on_error "false" \
  --misc.clean_on_exit true
```

:::tip
Using `go run` directly allows the version of `air` to be specified. This ensures that the version of `air` is consistent between machines. In addition, you don't need to run `air init` to generate `.air.toml`.
:::

:::note
This command doesn't do anything to restart or send a reload event to the `templ` proxy server. We'll use a separate `air` command to trigger a notify event when any non-go related files change.
:::

### Reload event

We also want the browser to automatically reload when the:

1. HTML content changes
2. CSS bundle changes
3. JavaScript bundle changes

To trigger the event, we can use the `air` command to use a different set of options, using the `templ` CLI to send a reload event to the browser.

```bash
go run github.com/cosmtrek/air@v1.51.0 \
  --build.cmd "templ generate --notify-proxy" \
  --build.bin "true" \
  --build.delay "100" \
  --build.exclude_dir "" \
  --build.include_dir "assets" \
  --build.include_ext "js,css"
```

:::note
The `build.bin` option is set to use the `true` command instead of executing the output of the `build.cmd` option, because the `templ generate --notify-proxy` command doesn't build anything, it just sends a reload event to the `templ` proxy server.

`true` is a command that exits with a zero status code, so you might see `Process Exit with Code 0` printed to the console.
:::

### Serving static assets

When using live reload, static assets must be served directly from the filesystem instead of being embedded in the Go binary, because the Go binary won't be re-built when the assets change.

In practice this means using `http.Dir` instead of `http.FS` to serve your assets.

If you don't want to do this, you can add additional asset file extensions to the `--build.include_ext` argument of the `air` command that rebuilds Go code to force a recompilation and restart of the Go server when the assets change.

#### Before

```go
//go:embed assets/*
var assets embed.FS
...
mux.Handle("/assets/", http.FileServer(http.FS(assets)))
```

#### After

```go
mux.Handle("/assets/", 
  http.StripPrefix("/assets", 
    http.FileServer(http.Dir("assets"))))
```

:::tip
Web browsers will cache assets when they receive a HTTP 304 response. This will result in asset changes not being visible within your application.

To avoid this, set the `Cache-Control` header to `no-store` for assets in development mode:

```go
var dev = true

func disableCacheInDevMode(next http.Handler) http.Handler {
	if !dev {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

mux.Handle("/assets/", 
  disableCacheInDevMode(
    http.StripPrefix("/assets", 
      http.FileServer(http.Dir("assets")))))
```
:::

## Putting it all together

A `Makefile` can be used to run all of the commands in parallel.

```make
# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	npx --yes tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch

# run esbuild to generate the index.js bundle in watch mode.
live/esbuild:
	npx --yes esbuild js/index.ts --bundle --outdir=assets/ --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

# start all 5 watch processes in parallel.
live: 
	make -j5 live/templ live/server live/tailwind live/esbuild live/sync_assets
```

:::note
The `-j5` argument to `make` runs all 5 commands in parallel.
:::

Run `make live` to start all of the watch processes.
