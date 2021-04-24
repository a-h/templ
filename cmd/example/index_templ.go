package templ

import "html"
import "io"
import "strings"
import strs "strings"

func Person(w io.Writer, p Person) error {
io.WriteString(w, "<div")
io.WriteString(w, " style=\"font-weight: bold\"")
io.WriteString(w, ">")
io.WriteString(w, html.EscapeString(p.Name()))
io.WriteString(w, "</div>")
io.WriteString(w, `
`)
return nil
}

