# ADR 0003: Ignore files for templ fmt and templ generate

## Status

Accepted

## Context

The `templ fmt` command walks directories and formats all `.templ` files it finds. With the addition of prettier attribute formatting (ADR 0002), `templ fmt` now modifies attribute values such as whitespace in class attributes and trailing semicolons in style values. Generator test fixtures (e.g. `generator/test-class-whitespace/`, `generator/test-element-attributes/`) contain intentionally messy attribute values to test the generator correctly. Running `templ fmt .` from the repository root modifies these fixtures, breaking the tests.

There was no mechanism to exclude directories or files from formatting or generation.

## Decision

Introduce two ignore files using gitignore-style glob syntax:

- `.templignore_fmt` -- patterns to skip during `templ fmt`
- `.templignore_generate` -- patterns to skip during `templ generate`

Each file is loaded from the root directory passed to the command (e.g. `templ fmt .` loads `./.templignore_fmt`). Lines are parsed with Go's `filepath.Match`: blank lines and `#`-prefixed comment lines are skipped, and the remaining lines are glob patterns. Matching checks the full relative path and each directory prefix, so a pattern like `generator/test-*` matches both `generator/test-foo` and `generator/test-foo/bar.templ`.

The implementation lives in `internal/ignorefile` with no new dependencies. The `Parse` function returns `nil` when the file does not exist, and `Matches` on nil patterns always returns false, so the feature is entirely opt-in with no behavior change when no ignore file is present.

## Consequences

- Generator test fixtures are protected from formatting changes without requiring manual intervention.
- Users can exclude vendor code, generated code, or other directories from formatting and generation.
- Stdin mode (`templ fmt < file.templ`) and LSP formatting are unaffected because they operate on individual files, not directory walks.
- Two separate files allow independent control over which paths are skipped for each command.

## Alternatives considered

1. **A single `.templignore` file for both commands.** Simpler, but users may want different exclusions for formatting vs. generation. Two files provide finer control.
2. **Command-line flags for exclude patterns.** This would work but requires remembering to pass the flags every time. A file in the repository is declarative and version-controlled.
3. **Respect `.gitignore` directly.** This would require a full gitignore parser (negative patterns, nested ignore files, `**` globstar). The simple `filepath.Match` approach covers the common cases without the complexity.
