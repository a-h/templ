# FAQ

## How can I migrate from templ version 0.1.x to templ 0.2.x syntax?

Versions of templ &lt;= v0.2.663 include a `templ migrate` command that can migrate v1 syntax to v2.

The v1 syntax used some extra characters for variable injection, e.g. `{%= name %}` whereas the latest (v2) syntax uses a single pair of braces within HTML, e.g. `{ name }`.
