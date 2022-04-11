// Code generated by templ@(devel) DO NOT EDIT.

package testcssusage

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "strings"

func green() templ.CSSClass {
	var templCSSBuilder strings.Builder
	templCSSBuilder.WriteString(`color:#00ff00;`)
	templCSSID := templ.CSSID(`green`, templCSSBuilder.String())
	return templ.ComponentCSSClass{
		ID: templCSSID,
		Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),
	}
}

func className() templ.CSSClass {
	var templCSSBuilder strings.Builder
	templCSSBuilder.WriteString(`background-color:#ffffff;`)
	templCSSBuilder.WriteString(string(templ.SanitizeCSS(`color`, red)))
	templCSSID := templ.CSSID(`className`, templCSSBuilder.String())
	return templ.ComponentCSSClass{
		ID: templCSSID,
		Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),
	}
}

func Button(text string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		// Element (standard)
		// Element CSS
		var var_1 templ.CSSClasses = templ.Classes(className(), templ.Class("&&&unsafe"), templ.SafeClass("safe"))
		err = templ.RenderCSS(ctx, w, var_1)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "<button")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = io.WriteString(w, " class=")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, templ.EscapeString(var_1.String()))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, " type=\"button\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, ">")
		if err != nil {
			return err
		}
		// StringExpression
		_, err = io.WriteString(w, templ.EscapeString(text))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</button>")
		if err != nil {
			return err
		}
		return err
	})
}

func ThreeButtons() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		// CallTemplate
		err = Button("A").Render(ctx, w)
		if err != nil {
			return err
		}
		// CallTemplate
		err = Button("B").Render(ctx, w)
		if err != nil {
			return err
		}
		// Element (standard)
		// Element CSS
		var var_2 templ.CSSClasses = templ.Classes(green())
		err = templ.RenderCSS(ctx, w, var_2)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "<button")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = io.WriteString(w, " class=")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, templ.EscapeString(var_2.String()))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, " type=\"button\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, ">")
		if err != nil {
			return err
		}
		// StringExpression
		_, err = io.WriteString(w, templ.EscapeString("Green"))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "</button>")
		if err != nil {
			return err
		}
		return err
	})
}

