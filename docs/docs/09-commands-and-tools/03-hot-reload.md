# Hot reload

To access a Go web application that uses templ in a web browser, a few things must happen:

1. `templ generate` must be executed, to create Go code (`*_templ.go` files) from the `*.templ` files.
2. The Go code must start a web server on a port, e.g. (`http.ListenAndServe("localhost:8080", nil)`.
3. The Go program must be ran, e.g. by running `go run .`.
4. The web browser must access or reload the page, e.g. `http://localhost:8080`. 

If the `*.templ` files change, #1 and #2 must be ran.

If the `*.go` files change, #3 and #4 must be ran.

## Built-in

templ ships with hot reload that carries out these operations.

templ uses a basic `os.WalkDir` function to iterate through `*.templ` files on disk, with a backoff strategy to prevent excessive disk thrashing and reduce CPU usage.

`templ generate --watch` watches the current directory for changes and will run `templ generate` if changes are detected (#1 and #2).

To re-run your app, set the `--cmd` argument, and templ will start or restart your app using the command provided once template code generation is complete (#3).

To trigger your web browser to reload automatically (without pressing F5), set the `--proxy` argument (#4).

The `--proxy` argument starts a HTTP proxy which proxies requests to your app. For example, if your app runs on port 8080, you would use `--proxy="http://localhost:8080"`. The proxy inserts client-side JavaScript before the `</body>` tag that will cause the browser to reload the window when the app is restarted instead of you having to reload the page manually.

Altogether, to setup hot reload on an app that listens on port 8080, run the following.

```
templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
```

The hot reload process can be shown in the following diagram:

```mermaid
sequenceDiagram
    browser->>templ_proxy: HTTP
    activate templ_proxy
    templ_proxy->>app: HTTP
    activate app
    app->>templ_proxy: HTML
    deactivate app
    templ_proxy->>templ_proxy: add reload script
    templ_proxy->>browser: HTML
    deactivate templ_proxy
    browser->>templ_proxy: SSE request to /_templ/reload/events
    activate templ_proxy
    templ_proxy->>generate: run templ generate if *.templ files have changed
    templ_proxy->>app: restart app if *.go files have changed
    templ_proxy->>browser: notify browser to reload page
    deactivate templ_proxy
```

## Alternative 1: wgo

[wgo](https://github.com/bokwoon95/wgo):

> Live reload for Go apps. Watch arbitrary files and respond with arbitrary commands. Supports running multiple invocations in parallel.

```
wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run main.go
```

To avoid a continous reloading files ending with `_templ.go` should be skipped via `-xfile`.

## Alternative 2: air

Air's reload performance is better than templ's built-in feature due to its complex filesystem notification setup, but doesn't ship with a proxy to automatically reload pages, and requires a `toml` configuration file for operation.

See https://github.com/cosmtrek/air for details.

### Example configuration

```toml title=".air.toml"
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "templ generate && go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor"]
  exclude_file = []
  exclude_regex = [".*_templ.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "templ", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```
