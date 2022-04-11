// Code generated by templ@(devel) DO NOT EDIT.

package testhtml

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"

func BasicTemplate(url string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		// Element (standard)
		_, err = io.WriteString(w, "<div>")
		if err != nil {
			return err
		}
		// Element (standard)
		// Element CSS
		// Element Script
		err = templ.RenderScripts(ctx, w, )
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "<a")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = io.WriteString(w, " href=")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		var var_1 templ.SafeURL = templ.URL(url)
		_, err = io.WriteString(w, templ.EscapeString(string(var_1)))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, ">")
		if err != nil {
			return err
		}
		// Text
		var_2 := `text`
		_, err = io.WriteString(w, var_2)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</a>")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</div>")
		if err != nil {
			return err
		}
		return err
	})
}

