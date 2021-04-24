package templ

import "html"
import "io"
import "strings"
import strs "strings"

func Person(w io.Writer, p Person) {
io.WriteString(w, html.EscapeString(p.Name()))
return nil
}

