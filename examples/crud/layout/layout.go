package layout

import (
	"net/http"

	"github.com/a-h/templ"
)

func Handler(content templ.Component) http.Handler {
	return templ.Handler(Page(content))
}
