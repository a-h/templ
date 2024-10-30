# Pass Go data to TypeScript

This demonstrates how to bundle TypeScript code, and use it in a templ project.

The TypeScript code is bundled using `esbuild`, with templ used to serve HTML.

The `importmap` used by the browser to resolve dependencies is generated using [jspm-cli](https://jspm.org/docs/jspm-cli/stable/).

The code demonstrates application of `onclick` event handlers, and how to pass data from Go to TypeScript.

This demo passes data from server-side Go code to TypeScript code by placing the data in `<script type="application/json">` tags, or data attributes (called `alert-data` in this example).

Note how the Go web server serves the `./assets` directory, which contains the bundled TypeScript code.

## Tasks

### generate

```bash
templ generate
```

### ts-install

Since it's a TypeScript project, you need to install the dependencies.

Dir: ts

```bash
npm install
```

### ts-build-cli

If you have `esbuild` installed globally, you can bundle and minify the TypeScript code without using NPM. Remember to run `npm install` to install the dependencies first.

```bash
esbuild --minify --outdir=assets/js ts/src/*index*.ts
```

### ts-build-npm

If you don't have `esbuild` installed globally, you can use the NPM script to build the TypeScript code.

Dir: ./ts

```bash
npm run build
```

### Add dependencies to the import-map

Dir: ./ts

```
npx jspm install --provider=jsdelivr --env=module lightweight-charts
```

### run

```bash
go run .
```
