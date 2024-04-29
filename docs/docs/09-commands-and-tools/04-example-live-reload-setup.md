# A Full Live Reload Setup

Browser live reload without manually clicking refresh is a great feature for developers. It allows you to see your changes immediately without having to switch to your browser and press `F5` or `CMD+R`. Projects usually involves multiple build processes, e.g. css bundling, js bundling, and Go compiling. Setting everything up can be a bit tricky.

Luckily combining a tool like `air` and `templ`'s built-in proxy server, you can get a full live reload setup. In this page, we'll setup a live reloading `Makefile` that covers:

- [Tailwindcss](https://tailwindcss.com/) for generating css bundle
- [esbuild](https://esbuild.github.io/) bundling Javascript or Typescript
- [air](https://github.com/cosmtrek/air) for re-building Go source as well as sending reload event to `templ` proxy server

## How does it work

The core capability for live reloading (i.e. automatically refreshes the browser when a file changes) is provided by `templ`'s built-in proxy server. The proxy server injects a bit of javascript to reload the browser when a "reload" event is sent down via SSE. See [Live Reload page](/commands-and-tools/live-reload) for detailed explanation.

:::tip
Importantly, this bit of javascript is only injected if your HTML file contains the closing `</body>` tag. If the live reloading stopped working, it could be the response doesn't contain the string `</body>`
:::

The reloading is done by running `templ generate --watch` which will send the "reload" event whenever a ".templ" file changes. The other way is to manually trigger it by sending a HTTP POST request to `/_templ/reload/event` endpoint. `templ` cli provides this via `templ generate --notify-proxy`

:::tip
templ proxy server `--watch` mode generates different `_templ.go` files. It creates `_templ.txt` files and serves that to the client directly. In other words, the `_templ.go` files don't change. This is to avoid unnecessary Go re-compilations.
:::

## Setting up the Makefile

### Templ watch mode

To start the `templ` proxy server in watch mode, run:

```shell
templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v
```

This assumes that your http server is running on `http://localhost:8080`. `--open-browser=false` is to prevent `templ` from opening the browser automatically. `-v` turns on verbose logging.

### Tailwindcss

Assuming you're have the correct setup for tailwindcss, i.e. having `tailwind.config.js` and `input.css` at the root of your project, you can run:

```shell
npx tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch
```

This will watch `input.css` as well as your `.templ` files and re-generate `styles.css` whenever there's a change.

### esbuild

If you have any javascript or typescript files that you want to bundle, you can use `esbuild`:

```shell
npx esbuild js/index.ts --bundle --outdir=assets/ --watch
```

This will watch `js/index.ts` and relevant files, and re-generate `assets/index.js` whenever there's a change.

### Re-build Go source

To watch and restart your Go server, you can use `air`:

```shell
go run github.com/cosmtrek/air@v1.51.0 \
  --build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
  --build.exclude_dir "node_modules" \
  --build.include_ext "go" \
  --build.stop_on_error "false" \
  --misc.clean_on_exit true
```

:::tip
Note that we're using `go run` directly so that we can specify the version of `air` to use. This is to ensure that the version of `air` is consistent across different machines. In addition, you don't need to run `air init` to generate `.air.toml`.
:::

This will watch your only your Go files and re-build the binary whenever there's a change.

Air is recommended for this task, because it handles killing the previous process correctly. Refer to [Air's documentation](https://github.com/cosmtrek/air?tab=readme-ov-file) for details of the command line arguments.

Clever readers might notice we didn't include anything to restart or send reload event to `templ` proxy server. This is because we'll use a separate `air` command to trigger notify event when any non-go relaed files changes.

### Reload event

If we summerise all the scenarios where we want the browser to automatically reload:

1. The html content changed
2. The css file changed
3. The javascript file changed

If any `.go` files changes, we just want it to be rebuilt and restarted when we click a button that requests that resource.

Therefore, our final `air` command will watch all the assets and send a reload event using `templ` cli:

```shell
go run github.com/cosmtrek/air@v1.51.0 \
  --build.cmd "templ generate --notify-proxy" \
  --build.bin "true" \
  --build.delay "100" \
  --build.exclude_dir "" \
  --build.include_dir "assets" \
  --build.include_ext "js,css"
```

Here we use a `true` command as the binary to build, because we don't want to build anything. We just want to send the reload event to the `templ` proxy server. You might see `Process Exit with Code 0` in the output, which is expected.

:::tip
In this setup, you should serve your static asset with http.Dir instead of using //go:embed and http.FS. E.g.

Instead of:
```go
//go:embed assets/*
var assets embed.FS
...
mux.Handle("/assets/", http.FileServer(http.FS(assets)))
```
Do:
```go
mux.Handle("/assets/", 
  http.StripPrefix("/assets", 
    http.FileServer(http.Dir("assets"))))
```

This is because you're not re-building Go binary when assets change, so the embedded assets won't be updated.

If you would like to use `//go:embed`, you can add the necessary extensions to `--build.include_ext` in the `air` command for **Rebuilding Go Source** section.
:::

## Putting it all together

You can put all the commands in a `Makefile`:

```makefile
# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Open http://localhost:7331
live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server
live/server:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle
live/tailwind:
	npx tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch

# run esbuild to generate the index.js bundle
live/esbuild:
	npx esbuild js/index.ts --bundle --outdir=assets/ --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ watch
live/sync_assets:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

# start all 5 watch processes
live: 
	make -j5 live/templ live/server live/tailwind live/esbuild live/sync_assets
```

Notice we added the `make -j5` command to run all the commands in parallel. This is to ensure that all the watch process runs in parallel.
