# Command-line tools

`templ` provides a command line interface. Most users will only need to run the `templ generate` command to generate Go code from `*.templ` files.

```
usage: templ <command> [parameters]
To see help text, you can run:
  templ generate --help
  templ fmt --help
  templ lsp --help
  templ migrate --help
  templ version
examples:
  templ generate
```

## Generating Go code from templ files

The `templ generate` command generates Go code from `*.templ` files in the current directory tree.

The command provides additional options:

```
  -cmd string
        Set the command to run after generating code.
  -f string
        Optionally generates code for a single file, e.g. -f header.templ
  -help
        Print help and exit.
  -path string
        Generates code for all files in path. (default ".")
  -pprof int
        Port to start pprof web server on.
  -proxy string
        Set the URL to proxy after generating code and executing the command.
  -proxyport int
        The port the proxy will listen on. (default 7331)
  -sourceMapVisualisations
        Set to true to generate HTML files to visualise the templ code and its corresponding Go code.
  -w int
        Number of workers to run in parallel. (default 4)
  -watch
        Set to true to watch the path for changes and regenerate code.
```

For example, to generate code for a single file:

```
templ generate -f header.templ
```

## Formatting templ files

The `templ fmt` command formats template files. You can use this command in different ways:

1. Format all template files in the current directory and subdirectories:

```
templ fmt .
```

2. Format input from stdin and output to stdout:

```
templ fmt
```

## Language Server for IDE integration

`templ lsp` provides a Language Server Protocol (LSP) implementation to support IDE integrations.

This command isn't intended to be used directly by users, but is used by IDE integrations such as the VSCode extension and by Neovim support.

A number of additional options are provided to enable runtime logging and profiling tools.

```
  -goplsLog string
        The file to log gopls output, or leave empty to disable logging.
  -goplsRPCTrace
        Set gopls to log input and output messages.
  -help
        Print help and exit.
  -http string
        Enable http debug server by setting a listen address (e.g. localhost:7474)
  -log string
        The file to log templ LSP output to, or leave empty to disable logging.
  -pprof
        Enable pprof web server (default address is localhost:9999)
```
