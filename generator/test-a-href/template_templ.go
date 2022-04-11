// Code generated by templ@(devel) DO NOT EDIT.

package testahref

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"

func render() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		// Element (standard)
		_, err = io.WriteString(w, "<a")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = io.WriteString(w, " href=\"javascript:alert(&#39;unaffected&#39;);\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, ">")
		if err != nil {
			return err
		}
		// Text
		var_1 := `Ignored`
		_, err = io.WriteString(w, var_1)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</a>")
		if err != nil {
			return err
		}
		// Element (standard)
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
		var var_2 templ.SafeURL = templ.URL("javascript:alert('should be sanitized')")
		_, err = io.WriteString(w, templ.EscapeString(string(var_2)))
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
		var_3 := `Sanitized`
		_, err = io.WriteString(w, var_3)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</a>")
		if err != nil {
			return err
		}
		// Element (standard)
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
		var var_4 templ.SafeURL = templ.SafeURL("javascript:alert('should not be sanitized')")
		_, err = io.WriteString(w, templ.EscapeString(string(var_4)))
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
		var_5 := `Unsanitized`
		_, err = io.WriteString(w, var_5)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</a>")
		if err != nil {
			return err
		}
		return err
	})
}

