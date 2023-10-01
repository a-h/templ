# Template generation

To generate Go code from `*.templ` files, use the `templ` command line tool.

```
templ generate
```

The `templ generate` command will recurse into subdirectories and generate Go code for every `*.templ` file it finds.

The command will output a list of files that it processed, how long it took, and the total elapsed time.

```
main.templ complete in 897.292µs
Generated code for 1 templates with 0 errors in 1.291292ms
```

## Advanced options

The `templ generate` command has a `--help` option that prints advanced options.

These include the ability to generate code for a single file and to choose the number of parallel workers that `templ generate` uses to create Go files.

By default `templ generate` uses the number of CPUs that your machine has installed.

```
templ generate --help
```

```
  -f string
        Optionally generates code for a single file, e.g. -f header.templ
  -help
        Print help and exit.
  -path string
        Generates code for all files in path. (default ".")
  -sourceMapVisualisations
        Set to true to generate HTML files to visualise the templ code and its corresponding Go code.
  -w int
        Number of workers to run in parallel. (default runtime.NumCPU())
```
