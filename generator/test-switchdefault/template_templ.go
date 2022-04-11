// Code generated by templ@(devel) DO NOT EDIT.

package testswitchdefault

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"

func template(input string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		// Text
		var_1 := `switch input `
		_, err = io.WriteString(w, var_1)
		if err != nil {
			return err
		}
		// StringExpression
		_, err = io.WriteString(w, templ.EscapeString(
		case "a":
			{ "it was 'a'" }
		default:
			{ "it was something else" }
	))
		if err != nil {
			return err
		}
		return err
	})
}

