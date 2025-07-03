# Coding standards

## Environment setup

templ has an `.envrc` file that is used to set up the development environment using a tool called `direnv`. There is a VS Code extension available that will automatically load this when you open the project in VS Code.

The `.envrc` file uses a Nix flake to set up the environment, so Nix is required to be installed.

The version of Go used is defined in the `flake.nix` file.

## Build tasks

templ uses the `xc` task runner - https://github.com/joerdav/xc

If you run `xc` you can get see a list of the development tasks that can be run, or you can read the `README.md` file and see the `Tasks` section.

The most useful tasks for local development are:

* `xc install-snapshot` - builds the templ CLI and installs it into `~/bin`. Ensure that this is in your path.
* `xc test` - regenerates all templates, and runs the unit tests.
* `xc fmt` - runs `gofmt` to format all Go code.
* `xc lint` - run the same linting as run in the CI process.
* `xc docs-run` - run the Docusaurus documentation site.

templ has a code generation step, this is automatically carried out using `xc test`.

## Commit messages

The project using https://www.conventionalcommits.org/en/v1.0.0/

Examples:

* `feat: support Go comments in templates, fixes #234"`

## Documentation

Documentation is written in Markdown, and is rendered using Docusaurus. The documentation is in the `docs` directory.

Update documentation when the behaviour of templ changes, or when new features are added.

## Coding style

* Reduce nesting - i.e. prefer early returns over an `else` block, as per https://danp.net/posts/reducing-go-nesting/ or https://go.dev/doc/effective_go#if
* Use line breaks to separate "paragraphs" of code - don't use line breaks in between lines, or at the start/end of functions etc.
* Use the `xc fmt` and `xc lint` build tasks to format and lint code before committing.
* Don't use unnecessary comments that explain what the code does.
* If comments are used, ensure that they are full sentences, and use proper punctuation, including ending with a full stop.
