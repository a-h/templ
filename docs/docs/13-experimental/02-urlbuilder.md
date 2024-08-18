# urlbuilder

A simple URL builder to construct a `templ.SafeURL`.

```templ title="component.templ"
import (
  "github.com/templ-go/x/urlbuilder"
  "strconv"
  "strings"
)

templ component(o Order) {
  <a
    href={ urlbuilder.New("https", "example.com").
    Path("orders").
    Path(o.ID).
    Path("line-items").
    Query("page", strconv.Itoa(1)).
    Query("limit", strconv.Itoa(10)).
    Build() }
  >
    { strings.ToUpper(o.Name) }
  </a>
}
```

See [URL Attribures](/syntax-and-usage/attributes#url-attributes) for more information.

## Feedback

Please leave your feedback on this feature at https://github.com/a-h/templ/discussions/867
