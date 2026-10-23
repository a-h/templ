# ADR 0008: Single prettier invocation per templ file

## Status

Accepted

## Context

Issue #1471: `templ fmt` became over six times slower in v0.3.1070. The regression was introduced by ADR 0002, which ran prettier once per `HTMLTemplate` block to format constant attribute values. Script and style elements already ran prettier once per element.

Starting prettier takes over 100ms, and the time is spent waiting on the child process rather than in Go code. A CPU profile of the formatter shows Go busy for under 1% of the wall time. ADR 0002 assumed that one extra invocation per template was the same order of magnitude as the existing calls, but a file with many small components gains one invocation per component, so the cost grows linearly with the number of templates.

Measured on an Apple M1 Max with prettier 3, formatting one file:

| Components | Invocations before | Time before | Invocations after | Time after |
|------------|--------------------|-------------|-------------------|------------|
| 1          | 3                  | 378ms       | 1                 | 313ms      |
| 10         | 12                 | 1447ms      | 1                 | 436ms      |
| 50         | 52                 | 6306ms      | 1                 | 525ms      |

## Decision

Collect every constant attribute, script element, and style element in a templ file, and format them in one synthetic HTML document with a single prettier invocation.

Each item in the document carries an index attribute (`data-templ-attribute-id`, `data-templ-script-id`, or `data-templ-style-id`) that correlates the formatted output back to the templ AST. Script and style elements are wrapped in `data-templ-depth` div elements, as before, so that prettier indents their contents to match their position in the templ file.

`TestPrettierBatchStartsPrettierOncePerFile` counts prettier invocations, and fails if formatting one file starts prettier more than once. The test fails instead of skipping when prettier is missing in CI, so the guard cannot disappear silently. CI also runs `xc benchmark-fmt` to record formatting times.

## Consequences

- Formatting time no longer grows with the number of templates or elements in a file.
- If prettier fails on one part of the file, no part of the file is formatted. Previously, the same failure also stopped formatting of the file, so the visible behavior does not change.
- The LSP formatter uses the same code path, so formatting on save is faster.

## Alternatives considered

1. **Batch attributes once per file, but keep one invocation per script and style element.** This fixes the regression, but a file with many script or style elements remains slow.
2. **Batch across files.** This reduces invocations further, but couples the formatting of unrelated files, so that one file's prettier error affects others. `templ fmt` already formats files in parallel.
3. **Assert a wall-clock time budget in CI.** CI runner performance varies, so a time budget either fails at random or is too loose to catch a regression. Counting invocations is deterministic and targets the cause of the cost.
