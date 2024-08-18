## Tasks

### build-templ

```
templ generate
```

### build-js

Dir: react

```
esbuild --bundle index.ts --outdir=../static --minify --global-name=bundle
```

### run

```
go run .
```

### all

Requires: build-templ
Requires: build-js
Requires: run

```
echo "Running"
```
