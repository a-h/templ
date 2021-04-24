package templ

import "html"
import "io"
import "strings"
import strs "strings"

func Address(w io.Writer, addr Address) error {
io.WriteString(w, `	`)
io.WriteString(w, "<div>")
io.WriteString(w, html.EscapeString(addr.Address1))
io.WriteString(w, "</div>")
io.WriteString(w, `
	`)
io.WriteString(w, "<div>")
io.WriteString(w, html.EscapeString(addr.Address2))
io.WriteString(w, "</div>")
io.WriteString(w, `
	`)
io.WriteString(w, "<div>")
io.WriteString(w, html.EscapeString(addr.Address3))
io.WriteString(w, "</div>")
io.WriteString(w, `
	`)
io.WriteString(w, "<div>")
io.WriteString(w, html.EscapeString(addr.Address4))
io.WriteString(w, "</div>")
io.WriteString(w, `
`)
return nil
}

func Person(w io.Writer, p Person) error {
io.WriteString(w, "<div")
io.WriteString(w, " style=\"font-weight: bold\"")
io.WriteString(w, " id=")
io.WriteString(w, "\"")
io.WriteString(w, html.EscapeString(p.ID))
io.WriteString(w, "\"")
io.WriteString(w, ">")
io.WriteString(w, html.EscapeString(p.Name()))
io.WriteString(w, "</div>")
io.WriteString(w, `

`)
for i, v := range p.Addresses{ 
io.WriteString(w, `	`)
Address(w, v)
io.WriteString(w, `
`)
}
io.WriteString(w, `

`)
if p.IsAdmin(){ 
io.WriteString(w, `	`)
io.WriteString(w, "<h1>")
io.WriteString(w, html.EscapeString("Admin"))
io.WriteString(w, "</h1>")
io.WriteString(w, `
`)
} else {
}
io.WriteString(w, `
`)
return nil
}

