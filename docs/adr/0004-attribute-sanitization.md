# ADR 0004: Sanitize attributes in the same way as html/template

## Status

Accepted

## Context

GHSA-94c2-xg24-pwhv reported that templ only sanitized `string` values in `<a href>`, `<link href>`, `<form action>`, and `<object data>`, and only when the attribute name was lowercase. Values in other attributes that execute `javascript:` URLs, such as `<iframe src>`, `<button formaction>`, `<area href>`, and `<a HREF>`, were HTML-escaped, but not sanitized. `<div OnClick={ s }>` accepted a `string`, because only lowercase `on*` attributes required a `templ.ComponentScript`.

The same class of issue affected:

- Attribute key expressions (`<a { name }={ value }>`), which were never sanitized, and whose names were HTML-escaped, but not validated, so a name such as `x onmouseover=alert(1)` added an event handler.
- Spread attributes (`<a { attrs... }>`), which were never sanitized, and had the same name validation issue.
- htmx event handler syntaxes other than `hx-on:`, i.e. `hx-on`, `hx-on-*`, `data-hx-on:*`, and `data-hx-on-*`.
- `srcdoc`, which the browser parses as an HTML document in the context of the page.
- SVG `<set>` and `<animate>` elements, which can set the `href` attribute of their parent element to a `javascript:` URL with the `to`, `from`, `values`, and `by` attributes.

## Decision

Classify each attribute by name, apply the same classification at generation time, for attributes with constant names, and at render time, for attribute key expressions and spread attributes, and treat the same attributes as URLs as Go's `html/template` package does.

`html/template` has had more security review than templ, and Go developers expect the same behavior from templ. Matching it means that templ's URL rules can be audited by comparison with `html/template`'s `attrType` function, rather than by reviewing a templ-specific list.

The classification lives in `internal/htmlattr`. `htmlattr.Classify` returns a `htmlattr.Context` for an element and attribute name, and the generator and the runtime each map the context to the escaping that it requires, in the same way that `parser.ScriptContentsContext` determines the escaping of Go expressions in `<script>` elements (GHSA-jq3r-rjqp-mwgj). A single classification function means that the generator and the runtime can't disagree on the rules. HTML element and attribute names are case insensitive, and browsers only lowercase the ASCII letters in them, so `Classify` does the same before comparing names.

- URL attributes, as in `html/template`, on any element: `action`, `archive`, `background`, `cite`, `classid`, `codebase`, `data`, `formaction`, `href`, `icon`, `longdesc`, `manifest`, `poster`, `profile`, `src`, `usemap`, and `xmlns`, the same names with a `data-` prefix or a namespace prefix, e.g. `xlink:href`, all `xmlns:*` attributes, and any attribute whose name contains `src`, `uri`, or `url`, e.g. `data-url`. Values are sanitized with `templ.URL`, unless they are a `templ.SafeURL`.
- Event handler attributes: names starting with `on`, `hx-on`, or `data-hx-on`. Constant names require a `templ.ComponentScript` at compile time. At render time, values that are not a `templ.ComponentScript` are replaced with `templ.FailedSanitizationJS`.
- `srcdoc`, as in `html/template`, on any element, and `data-srcdoc`: values are escaped as text, unless they are a `templ.Component`, which is rendered as HTML.
- SVG `<set>` and `<animate>` `to`, `from`, `values`, and `by`: each item in the semicolon separated list is sanitized with `templ.URL`, unless the element has a constant `attributeName` attribute that can't refer to `href`, and that attribute appears before any spread, conditional, or dynamic attribute. `html/template` doesn't sanitize these attributes.
- Attribute names that are only known at render time are written to the output as they are, so they can't contain characters that end the name, or start a new attribute, i.e. whitespace, control characters, `"`, `'`, `` ` ``, `<`, `>`, `/`, and `=`. Names that contain `&` are also rejected. In an attribute value, templ escapes `&` as `&amp;`, and the browser decodes it back to `&`, but browsers don't decode character references in attribute names, so escaping `&` in a name would produce an attribute with a different name. Rejected names are replaced with `templ.FailedSanitizationAttributeName`.

Render time sanitization replaces unsafe values, in the same way as `templ.FailedSanitizationURL`, rather than returning an error.

Generated code passes the result of each attribute expression to `templruntime.NewAttributeExpression`, and resolves it with a method that takes the context, e.g. `templruntime.NewAttributeExpression(p.URL).ResolveURL(ctx)`. The constructor accepts the same types as other attribute values, e.g. `int`, and an optional error, so expressions such as `{ getURL() }` that return a value and an error work in any attribute. Passing the context means that templ can log or record sanitization in future without changing the generated code.

The functions and types that generated code uses are in the `runtime` package, imported as `templruntime`, so that the `templ` package only contains functions and types for users. `templ.RenderAttributes`, which generated code used for spread attributes, is removed rather than deprecated. It can't call the `runtime` package without an import cycle, and leaving it in place would leave spread attributes unsanitized in code generated by earlier versions of templ. Removing it means that code generated by an earlier version fails to compile, so that the vulnerability is fixed by regenerating, which updating templ requires anyway.

Fuzz tests in `runtime/fuzzing` render fuzzed attribute names and values through each path, parse the output with an HTML5 parser, and assert that no attribute was added, and that no value can execute JavaScript. The assertion of what a browser executes is written independently of `htmlattr.Classify`, so that a missing classification fails the fuzz tests, rather than changing what they expect.

## Consequences

- `string` values in all attributes that `html/template` treats as URLs are sanitized, including `<script src>`, so a `data:text/javascript` URL can't be used to load a script.
- `data:` and `blob:` URLs fail sanitization in all URL attributes, including `<img src>`, so inline images require `templ.SafeURL`, in the same way that `html/template` requires `template.URL`.
- URL attributes are sanitized on custom elements, e.g. `<turbo-stream action>`. Values without a scheme, e.g. `append`, are unchanged.
- `srcset` contains `src`, so templ sanitizes it as a single URL, while `html/template` sanitizes each URL in it separately. A `srcset` that contains a `data:` URL fails sanitization.
- `<div OnClick={ s }>` and `<div hx-on-click={ s }>` no longer compile when `s` is a `string`.
- `templ.Attributes{"onclick": "alert(1)"}` renders `onclick="/* templ: failed sanitization, use templ.JSFuncCall, or templ.JSUnsafeFuncCall for trusted JavaScript */"`.
- `string` values in `srcdoc` attributes are displayed as text.
- The JavaScript-evaluating attributes of other frameworks, such as Alpine.js and Datastar, are not sanitized, and are documented as the responsibility of the developer. This includes htmx attributes other than event handlers, e.g. `hx-vals` and `hx-headers` values with a `js:` prefix, `hx-vars`, and `hx-trigger` event filters. htmx has too many attributes and extensions that execute JavaScript for templ to sanitize a fixed list of them, so templ keeps its handling of htmx event handler attributes, and documents the rest.
- `templ.RenderAttributes` is removed. Code generated by an earlier version of templ fails to compile until it is regenerated.
- `templ.ResolveAttributeValue` and `templ.JoinURLErrs` are deprecated, and will be removed in favour of `templruntime.NewAttributeExpression`. Code generated by an earlier version of templ that uses them continues to compile, but does not sanitize URLs in attributes such as `<iframe src>` until it is regenerated.

## Alternatives considered

1. **Only sanitize attributes that can execute JavaScript in the page.** An earlier version of this fix sanitized `href`, `xlink:href`, and `formaction` on any element, `<form action>`, `<object data>`, and `src` on `<iframe>`, `<frame>`, and `<embed>`, so that `data:` image URLs in `<img src>` continued to work. This left `<script src>` unsanitized, where a `data:text/javascript` URL executes, and custom attributes such as `data-url` unsanitized, where JavaScript on the page may navigate to the URL. It also required templ to maintain its own judgement of which attributes are dangerous, which has had less review than `html/template`'s.
2. **Return an error at render time for unsafe values.** This makes unsafe values visible, but turns an existing page that renders into a page that fails to render. Replacing the value is consistent with `templ.FailedSanitizationURL`, and the replacement value is visible in the HTML.
3. **A denylist of `javascript:` and `vbscript:` schemes, instead of the `templ.URL` allowlist.** This would allow `data:` and `blob:` URLs, but browsers ignore whitespace and control characters within schemes, so a denylist requires normalization that is difficult to get right. The allowlist is unchanged.
4. **Require `templ.ComponentScript` for the attributes of other JavaScript frameworks.** This protects more templates, but couples templ to third-party attribute syntax, and breaks more existing templates.
5. **The `html/template` attribute name filter.** `html/template` only allows ASCII letters and digits in a dynamic attribute name, and rejects names of URL, JavaScript, and CSS attributes entirely. This would replace common names such as `data-testid`, which templ documents as a use of attribute key expressions.
