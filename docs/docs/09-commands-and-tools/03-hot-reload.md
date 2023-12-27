# Hot reload

## Built-in

templ ships with hot reload. Since `fsnotify` and `rjeczalik/notify` filesystem watchers struggle to provide a working cross-platform behaviour, templ uses a basic `os.WalkDir` function to iterate through `*.templ` files on disk, and uses a backoff strategy to prevent excessive disk thrashing and reduce CPU usage.

`templ generate --watch` will watch the current directory and will templ files if changes are detected.

If the `--cmd` argument is set, templ start or restart the command once template code generation is complete.

If the `--proxy` argument is set, templ will start a HTTP proxy pointed at the given address. The proxy rewrites HTML received from the given address and adds a script just before the `</body>` tag that will reload the window with JavaScript once the changes are complete and the command has been executed.

```
templ generate --watch --proxy="http://localhost:8080" --cmd="runtest"
```

## Alternative

Air's reload performance is better due to its complex filesystem notification setup, but doesn't ship with a proxy to automatically reload pages, and requires a `toml` configuration file for operation.

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
